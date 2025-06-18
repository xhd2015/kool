package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

const launchConfig = `{
    "configurations": [
        {
            "name": "Debug dlv localhost:2345",
            "type": "go",
            "debugAdapter": "dlv-dap",
            "request": "attach",
            "mode": "remote",
            "port": 2345,
            "host": "127.0.0.1",
            "cwd":"./"
        },
        {
            "name": "Launch src/cmd/promq",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "program": "./src/cmd/promq",
            "cwd": "./",
            "args": [
                "get-db-qps",
                "--group-by",
                "service",
                "--endpoint",
                "biz-id",
                "--start",
                "2024-11-11T01:10:00+08:00",
                "--end",
                "2024-11-11T01:11:00+08:00"
            ]
        },
        {
            "name": "[ph-uat] Launch debug-main.bin",
            "type": "go",
            "request": "launch",
            "mode": "exec",
            "program": "./debug-main.bin",
            "cwd": "${workspaceFolder}",
            "env": {
                "GOLANG_PROTOBUF_REGISTRATION_CONFLICT": "ignore",
                "JAEGER_AGENT_PORT": "9999",
                "HOST_IP": "10.12.160.133",
                "LOCAL_DISABLE_INIT_BLOOM_FILTER_START": "true",
                "LOCAL_DISABLE_REQUEST_TIMEOUT": "true",
                "LOCAL_DISABLE_ANTI_FRAUD_CHECK": "true",
                "LOCAL_DISABLE_UID_ALLOCATOR": "true",
                "LOCAL_DISABLE_REGISTER_ETCD": "true",
                "HTTP_SERVER_PORT": "15001",
                "POD_IP": "xhd.localhost",
                "REGION": "ph",
                "ENV": "uat",
                // "GO_TEST_PROHIBIT_ARGS": "true",
            },
            "args": [
                "--config_dir=./config/ph/uat",
                "--config_file=config.ini",
                "--logger_file=logger.ini",
                "--sharding_file=sharding.ini"
            ],
            "preLaunchTask": "with-go1.19 go build -o debug-main.bin -mod=vendor ./src"
        }
    ]
}`

const tasks = `{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "Create dev terminals",
            "dependsOn": [
                "debug-main.bin",
                "bun run watch",
            ],
            // Mark as the default build task so cmd/ctrl+shift+b will create them
            "group": {
                "kind": "build",
                "isDefault": true
            },
            "runOptions": {}
        },
        {
            "label": "with-go1.19 go build -o debug-main.bin -mod=vendor ./src",
            "type": "shell",
            "command": "bash --login -i <<<'with-go1.19 go build  -mod=vendor -gcflags=all=\"-N -l\" -o ./debug-main.bin ./src'",
            "options": {
                "cwd": "${workspaceFolder}",
            }
        },
        {
            "type": "process",
            "label": "debug-main.bin",
            "command": "bash",
            "args": [
                "-c",
                "echo './debug-main.bin' | bash --login -i"
            ],
            "isBackground": true,
            "runOptions": {},
            "presentation": {
                "group": "dev"
            },
            "dependsOn": [
                "with-go1.19 go build -o debug-main.bin -mod=vendor ./src",
            ]
        },
        {
            "type": "process",
            "label": "bun run watch",
            "command": "bash",
            "args": [
                "-c",
                "echo 'cd frontend && bun run watch' | bash --login -i"
            ],
            "isBackground": true,
            "runOptions": {},
            "presentation": {
                "group": "dev"
            },
            "dependsOn": []
        }
    ]
}`

const createTaskTemplate = `{
    "version": "2.0.0",
    "tasks": [
        {
            "label": "Create dev terminals",
            "dependsOn": [
                "__CMD_NAME__",
            ],
            // Mark as the default build task so cmd/ctrl+shift+b will create them
            "group": {
                "kind": "build",
                "isDefault": true
            },
            "runOptions": {}
        },
        {
            "label": "__CMD_NAME__",
            "type": "process", // available "shell"
            "command": "bash",
            "args": [
                "-c",
                "echo '__CMD_ARGS__' | bash --login -i"
            ],
            "options": {
                "cwd": "${workspaceFolder}",
            }
        }
    ]
}`

func handleVscode(args []string) error {
	if len(args) > 0 {
		switch args[0] {
		case "debug-go":
			return handleVscodeDebugGo(args[1:])
		case "create-task":
			return handleVscodeCreateTask(args[1:])
		default:
			return fmt.Errorf("unrecognized command: %s", args[0])
		}
	}
	fmt.Printf(".vscode/launch.json\n%s\n", launchConfig)
	fmt.Printf(".vscode/tasks.json\n%s\n", tasks)
	return nil
}

func handleVscodeCreateTask(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("requires extra command")
	}
	cmd0 := args[0]
	if len(args) > 1 {
		return fmt.Errorf("unrecognized extra command: %s", strings.Join(args[1:], " "))
	}
	extraCommand := cmd0
	taskConfig := strings.ReplaceAll(createTaskTemplate, "__CMD_NAME__", extraCommand)
	taskConfig = strings.ReplaceAll(taskConfig, "__CMD_ARGS__", extraCommand)
	fmt.Printf(".vscode/tasks.json\n%s\n", taskConfig)
	return nil
}

func handleVscodeDebugGo(args []string) error {
	remainArgs := args[1:]
	if len(remainArgs) == 0 {
		return fmt.Errorf("requires <program>\nusage: kool vscode debug-go [--dlv] <program> [args...]")
	}

	var debugConfTemplate string
	var formatConf func(prog string, progArgs []string) (string, string)
	if remainArgs[0] == "--dlv" {
		debugConfTemplate = `{
"name": "Debug dlv localhost:2345",
"type": "go",
"debugAdapter": "dlv-dap",
"request": "attach",
"mode": "remote",
"port": 2345,
"host": "127.0.0.1",
"cwd":"./",
"preLaunchTask": "check dlv ready on localhost:2345"
}`
		remainArgs = remainArgs[1:]
		if len(remainArgs) == 0 {
			return fmt.Errorf("requires <program>\nusage: kool vscode debug-go --dlv <program> [args...]")
		}
		taskConfigTemplate := `{
"version": "2.0.0",
"tasks": [
    {
        "label": "go build -o __debug-main.bin ./",
        "type": "shell",
        "command": "bash --login -i <<<'go build -gcflags=all=\"-N -l\" -o ./__debug-main.bin ./'",
        "options": {
            "cwd": "${workspaceFolder}",
        }
    },
    {
        "label": "dlv exec --listen=localhost:2345",
        "type": "shell",
        "isBackground": true,
        "command": "bash --login -i <<<'dlv exec --api-version=2 --listen=localhost:2345  --headless ./__debug-main.bin%s >dlv.log 2>&1 &'",
        "options": {
            "cwd": "${workspaceFolder}",
        },
        "dependsOn": [
            "go build -o __debug-main.bin ./",
        ]
    },
    {
        "label": "check dlv ready on localhost:2345",
        "type": "shell",
        "command": "bash --login -i <<<'while true;do if grep -q \"API server listening at\" dlv.log;then echo ready; exit;fi ;sleep 1;done'",
        "options": {
            "cwd": "${workspaceFolder}",
        },
        "dependsOn": [
            "dlv exec --listen=localhost:2345",
        ]
    }
]
}`

		formatConf = func(prog string, progArgs []string) (debugConf string, taskConfig string) {
			debugConf = debugConfTemplate
			taskArg := ""
			if len(progArgs) > 0 {
				// TODO: proper quote
				taskArg = " -- " + strings.Join(progArgs, " ")
			}
			taskConfig = fmt.Sprintf(taskConfigTemplate, taskArg)
			return
		}

	} else {
		debugConfTemplate = `{
"name": "Launch Package",
"type": "go",
"request": "launch",
"mode": "auto",
"program": %q,
"cwd": "${workspaceFolder}",
"args": %s,
"env":{
    // "GOROOT":"goXXX"
    // "PATH":"goXXX/bin:${env:PATH}"
}
}`
		formatConf = func(prog string, progArgs []string) (string, string) {
			var argJSON string = "[]"
			if len(progArgs) > 0 {
				argJSONData, err := json.MarshalIndent(progArgs, "", "  ")
				if err != nil {
					panic(err)
				}
				argJSON = string(argJSONData)
			}
			debugConf := fmt.Sprintf(debugConfTemplate, prog, argJSON)
			return debugConf, ""
		}
	}

	prog := remainArgs[0]
	progArgs := remainArgs[1:]

	debugConf, taskConf := formatConf(prog, progArgs)
	exampleConfig := fmt.Sprintf(`{
// Use IntelliSense to learn about possible attributes.
// Hover to view descriptions of existing attributes.
// For more information, visit: https://go.microsoft.com/fwlink/?linkid=830387
"version": "0.2.0",
"configurations": [
%s
]
}`, IndentLines(debugConf, "        "))
	fmt.Printf(".vscode/launch.json\n%s\n", exampleConfig)
	if taskConf != "" {
		fmt.Printf(".vscode/tasks.json\n%s\n", taskConf)
	}
	return nil

}

func IndentLines(content string, prefix string) string {
	lines := strings.Split(content, "\n")
	n := len(lines)
	for i := 0; i < n; i++ {
		lines[i] = prefix + lines[i]
	}
	return strings.Join(lines, "\n")
}
