# Errno Code Gen
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