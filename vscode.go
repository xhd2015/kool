package main

import "fmt"

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
                "CAMPAIGN_GUARDIAN_ALLOW_DEBUG": "true",
                "CAMPAIGN_GUARDIAN_ALLOW_TODO_REMOVE": "true",
                "CAMPAIGN_GUARDIAN_DEBUG_SYNC_SCAN_LINK": "true",
                "CAMPAIGN_GUARDIAN_DISABLE_TASK": "true"
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
            "label": "with-go1.19 go build -o debug-main.bin -mod=vendor ./src",
            "type": "shell",
            "command": "bash --login -i <<<'with-go1.19 go build  -mod=vendor -gcflags=all=\"-N -l\" -o ./debug-main.bin ./src'",
            "options": {
                "cwd": "${workspaceFolder}",
            }
        }
    ]
}`

func handleVscode(args []string) error {
	fmt.Printf(".vscode/launch.json\n%s\n", launchConfig)
	fmt.Printf(".vscode/tasks.json\n%s\n", tasks)
	return nil
}