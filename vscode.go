package main

import (
	"encoding/json"
	"fmt"
)

const launchConfig = `{
    "configurations": [
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

func handleVscode(args []string) error {
	if len(args) > 0 && args[0] == "debug-go" {
		remainArgs := args[1:]
		if len(remainArgs) == 0 {
			return fmt.Errorf("requires <program>\nusage: kool vscode debug-go <program> [args...]")
		}
		prog := remainArgs[0]
		progArgs := remainArgs[1:]

		var argJSON string = "[]"
		if len(progArgs) > 0 {
			argJSONData, err := json.MarshalIndent(progArgs, "", "  ")
			if err != nil {
				return err
			}
			argJSON = string(argJSONData)
		}
		exampleConfig := fmt.Sprintf(`{
    // Use IntelliSense to learn about possible attributes.
    // Hover to view descriptions of existing attributes.
    // For more information, visit: https://go.microsoft.com/fwlink/?linkid=830387
    "version": "0.2.0",
    "configurations": [
        {
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
        }
    ]
}`, prog, argJSON)
		fmt.Printf(".vscode/launch.json\n%s\n", exampleConfig)
		return nil
	}
	fmt.Printf(".vscode/launch.json\n%s\n", launchConfig)
	fmt.Printf(".vscode/tasks.json\n%s\n", tasks)
	return nil
}
