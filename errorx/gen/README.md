# Errno Code Gen
## Production

When use this tool generating errno code, you should create two file at least for system metadata, such as: `./example/example_metadata.yaml` and `./example/example_common.yaml`.
It will provide the project errno settings and common error code, and then you should create a file for every biz module like `./example/example_user.yaml`.

## Usage
```
go run code_gen.go \
     --biz {bizName} \
     --app-name {appName} \
     --app-code {app-code} \
     --import-path {import-path} \
     --output-dir {output-dir} \
     --script-dir {script-dir}
```

## Example
Before use this command, you should rename or touch a new file without "example" prefix in example directory:
```
go run code_gen.go \
     --biz user \
     --app-name myapp \
     --app-code 6 \
     --import-path "github.com/crazyfrankie/frx/errorx/code" \
     --output-dir "./generated/user" \
     --script-dir "./example"
```