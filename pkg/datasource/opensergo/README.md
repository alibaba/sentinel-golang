# Opensergo DataSource for Sentinel Go

## Prepare Environment for Test Version

- Prepare the `OpenSergo GO SDK`.
  Because of the `OpenSergo GO SDK` has no published version, so should download the sourcecode of [`OpenSergo GO SDK`](https://github.com/jnan806/opensergo-go-sdk/tree/initial-version), and move it into you `GOPATH`.
- Rename the right version in directory name of `OpenSergo GO SDK` sourcecode. 
  Make sure the version in sourcecode directory name is the same with go.mod.  
  eg. `$PATH/pkg/mod/github.com/opensergo/opensergo-go@v0.0.0-20220331070310-e5b01fee4d1c`

## Samples

- [datasource_opensergo_example.go](./demo/datasource_opensergo_example.go)