# gitHook

Small webhook CICD app to deploy ALT-F4 site.

## config
config file is loacted at `~/.config/githook/config.yml`
and you can configure following parameters
```yml
token: <WebHook Token>
url: <base_url>
command: <command to run whin required event detected>
port: <port number where service will listen>
```