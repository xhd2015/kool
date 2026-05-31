package create

import (
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/less-flags"
	"github.com/xhd2015/xgo/support/cmd"
)

//go:embed all:go_react
var goReactTemplateFS embed.FS

const goReactHelp = `	
Usage: kool create go-react [--go-module <module-path>] <project-name>

Create a new go-react project.

  --go-module  specify the Go module path (e.g. github.com/user/repo)
               otherwise auto-detected from git remote, falls back to <project-name>
`

func HandleCreateGoReact(args []string) error {
	var goModule string
	args, err := lessflags.String("--go-module", &goModule).
		Help("-h,--help", goReactHelp).
		Parse(args)
	if err != nil {
		return err
	}

	if len(args) == 0 {
		return fmt.Errorf("requires project, try `kool create go-react --help`")
	}
	projectDir := filepath.Clean(args[0])
	args = args[1:]

	if len(args) > 0 {
		return fmt.Errorf("unrecognized extra arguments: %v", strings.Join(args, ","))
	}

	_, err = prepareProjectDir(projectDir)
	if err != nil {
		return err
	}

	err = copyTemplateFile(goReactTemplateFS, "go_react/.gitignore", filepath.Join(projectDir, ".gitignore"), "", "")
	if err != nil {
		return err
	}

	baseProjectName := filepath.Base(projectDir)

	engine := "bun"

	reactProjectName := baseProjectName + "-react"
	err = cmd.Debug().Dir(projectDir).Run(engine, "create", "vite", reactProjectName, "--template", "react-ts")
	if err != nil {
		return err
	}
	reactDir := filepath.Join(projectDir, reactProjectName)
	err = cmd.Debug().Dir(reactDir).Run("bun", "install")
	if err != nil {
		return err
	}

	// Install react-router-dom
	err = cmd.Debug().Dir(reactDir).Run("bun", "add", "react-router-dom")
	if err != nil {
		return err
	}

	// Install Node types for vite.config.ts aliases that use node:* imports.
	err = cmd.Debug().Dir(reactDir).Run("bun", "add", "-d", "@types/node")
	if err != nil {
		return err
	}

	// Update App.css
	err = updateAppCSS(filepath.Join(reactDir, "src", "App.css"))
	if err != nil {
		return fmt.Errorf("failed to update App.css: %v", err)
	}

	// Update index.css
	err = updateIndexCSS(filepath.Join(reactDir, "src", "index.css"))
	if err != nil {
		return fmt.Errorf("failed to update index.css: %v", err)
	}

	// Older vite templates produced `public/vite.svg`; newer ones (>=8.x)
	// produce `public/favicon.svg` instead. Handle whichever is present so
	// the scaffold keeps working across vite versions.
	err = renameViteIcon(reactDir, baseProjectName)
	if err != nil {
		return err
	}

	// Initialize Go Module
	modulePath := goModule
	if modulePath == "" {
		modulePath, _ = suggestGoModPath(projectDir)
	}
	if modulePath == "" {
		modulePath = baseProjectName
	}

	err = cmd.Debug().Dir(projectDir).Run("go", "mod", "init", modulePath)
	if err != nil {
		return err
	}

	// Copy Backend Template Files
	backendRoot := "go_react/backend"
	err = copyTemplateDir(goReactTemplateFS, backendRoot, projectDir, baseProjectName, modulePath)
	if err != nil {
		return err
	}

	// Rename existing App.tsx to AppGen.tsx before copying frontend templates
	appTsx := filepath.Join(reactDir, "src", "App.tsx")
	appGenTsx := filepath.Join(reactDir, "src", "AppGen.tsx")

	hasAppGen := false
	if _, err := os.Stat(appTsx); err == nil {
		err = os.Rename(appTsx, appGenTsx)
		if err != nil {
			return fmt.Errorf("failed to rename App.tsx to AppGen.tsx: %v", err)
		}
		hasAppGen = true
	}

	// Copy Frontend Template Files
	frontendRoot := "go_react/frontend"
	err = copyTemplateDir(goReactTemplateFS, frontendRoot, reactDir, baseProjectName, modulePath)
	if err != nil {
		return err
	}

	if err := initGitRepo(filepath.Join(reactDir, "external_src")); err != nil {
		return err
	}

	// Post-process App.tsx to handle AppGen conditional
	err = processAppTsx(filepath.Join(reactDir, "src", "App.tsx"), hasAppGen)
	if err != nil {
		return fmt.Errorf("failed to process App.tsx: %v", err)
	}

	// Post-process AppGen.tsx to fix logo reference
	if hasAppGen {
		err = processAppGenTsx(appGenTsx)
		if err != nil {
			return fmt.Errorf("failed to process AppGen.tsx: %v", err)
		}
	}

	err = cmd.Debug().Dir(reactDir).Run("bun", "run", "build")
	if err != nil {
		return err
	}

	// Go Mod Tidy
	err = cmd.Debug().Dir(projectDir).Run("go", "mod", "tidy")
	if err != nil {
		return err
	}

	// Git Init
	if err := initGitRepo(projectDir); err != nil {
		return err
	}
	return nil
}

func copyTemplateDir(templateFS embed.FS, srcRoot, targetDir, projectName, moduleName string) error {
	return fs.WalkDir(templateFS, srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		targetFilePath := filepath.Join(targetDir, relPath)

		if d.IsDir() {
			return os.MkdirAll(targetFilePath, 0755)
		}

		contentBytes, err := templateFS.ReadFile(path)
		if err != nil {
			return err
		}
		content := string(contentBytes)

		// Replace placeholders
		content = strings.ReplaceAll(content, "PROJECT_NAME", projectName)
		content = strings.ReplaceAll(content, "MODULE_NAME", moduleName)

		return os.WriteFile(targetFilePath, []byte(content), 0644)
	})
}

func copyTemplateFile(templateFS embed.FS, src, target, projectName, moduleName string) error {
	contentBytes, err := templateFS.ReadFile(src)
	if err != nil {
		return err
	}
	content := string(contentBytes)
	content = strings.ReplaceAll(content, "PROJECT_NAME", projectName)
	content = strings.ReplaceAll(content, "MODULE_NAME", moduleName)
	return os.WriteFile(target, []byte(content), 0644)
}

func replaceFile(f string, old, new string) error {
	content, err := os.ReadFile(f)
	if err != nil {
		return err
	}
	content = []byte(strings.Replace(string(content), old, new, 1))
	return os.WriteFile(f, content, 0644)
}

// renameViteIcon renames the vite-scaffolded icon (either `vite.svg` or
// newer `favicon.svg`) in <reactDir>/public to <projectName>.svg, and
// updates the corresponding reference in <reactDir>/index.html.
//
// If neither icon is present (e.g. a future vite template), it returns
// nil so scaffolding can still proceed; the project's index.html will
// simply keep its original href.
func renameViteIcon(reactDir, projectName string) error {
	publicDir := filepath.Join(reactDir, "public")
	target := projectName + ".svg"

	candidates := []string{"vite.svg", "favicon.svg"}
	var source string
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(publicDir, c)); err == nil {
			source = c
			break
		}
	}
	if source == "" {
		return nil
	}
	if source == target {
		return nil
	}

	err := os.Rename(filepath.Join(publicDir, source), filepath.Join(publicDir, target))
	if err != nil {
		return fmt.Errorf("rename %s -> %s: %v", source, target, err)
	}

	return replaceFile(filepath.Join(reactDir, "index.html"), "/"+source, "/"+target)
}

func updateAppCSS(path string) error {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(contentBytes), "\n")

	var newLines []string
	inRootBlock := false
	rootBlockModified := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "#root {") {
			inRootBlock = true
			newLines = append(newLines, line)
			if !rootBlockModified {
				newLines = append(newLines, "  width: 100%;")
				newLines = append(newLines, "  height: 100%;")
				rootBlockModified = true
			}
			continue
		}

		if inRootBlock {
			if strings.Contains(line, "max-width") {
				continue
			}
			if strings.Contains(line, "padding") {
				continue
			}
			if strings.HasPrefix(trimmed, "}") {
				inRootBlock = false
			}
		}

		newLines = append(newLines, line)
	}

	return os.WriteFile(path, []byte(strings.Join(newLines, "\n")), 0644)
}

func updateIndexCSS(path string) error {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(contentBytes), "\n")

	var newLines []string
	inBodyBlock := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "body {") {
			inBodyBlock = true
			newLines = append(newLines, line)
			continue
		}

		if inBodyBlock {
			if strings.Contains(line, "place-items") {
				continue
			}
			if strings.HasPrefix(trimmed, "}") {
				inBodyBlock = false
			}
		}

		newLines = append(newLines, line)
	}

	return os.WriteFile(path, []byte(strings.Join(newLines, "\n")), 0644)
}

func processAppTsx(path string, hasAppGen bool) error {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(contentBytes)

	if hasAppGen {
		content = strings.ReplaceAll(content, "__APP_GEN_IMPORT__", "import AppGen from './AppGen';")
		content = strings.ReplaceAll(content, "__APP_GEN_LINK__", "<Link to=\"/gen\">Generated App</Link>")
		content = strings.ReplaceAll(content, "__APP_GEN_ROUTE__", "<Route path=\"/gen\" element={<AppGen />} />")
	} else {
		content = strings.ReplaceAll(content, "__APP_GEN_IMPORT__", "")
		content = strings.ReplaceAll(content, "__APP_GEN_LINK__", "")
		content = strings.ReplaceAll(content, "__APP_GEN_ROUTE__", "")
	}

	return os.WriteFile(path, []byte(content), 0644)
}

func processAppGenTsx(path string) error {
	contentBytes, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines := strings.Split(string(contentBytes), "\n")

	var newLines []string
	for _, line := range lines {
		if strings.Contains(line, "import viteLogo") {
			continue
		}
		line = strings.ReplaceAll(line, "src={viteLogo}", "src={reactLogo}")
		newLines = append(newLines, line)
	}

	return os.WriteFile(path, []byte(strings.Join(newLines, "\n")), 0644)
}
