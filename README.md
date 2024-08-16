# cert_bot (Custom version)
The purpose of this branch is to set default `domains` and combine `dns01.json`, `account.json` into executable binary at the build stage.

Suggest using this repo as template, then make it private if needed.

# How to build(Manually)
See `.github/workflows/custom.yml` for details.
+ Set env `domains` as input value `-domains`(Optional)
+ Set valid `dns01.json`(Optional, let it be empty if you don't want to)
+ Set valid `account.json`(Optional, let it be empty if you don't want to)
+ Setup go environment
  ```sh
  go mod download
  go install github.com/rakyll/statik@v0.1.7
  go generate
  ```
+ `make -j build`

# How to build(Using Github Actions)
Just manually trigger the workflow dispatch of `custom` branch.

**Notice**: If the release asset with the same name exists, if will fail to upload!!! 


# What's the difference between custom & normal version
+ The custom one set `domains` you set as default, while the normal empty as default.
+ The custom one find `dns01.json` in the disk refer to cmd args input first, secondary find it in memory, else behaves like the normal one.
+ The custom one find `account.json` in the disk refer to cmd args input first, secondary find it in memory, else behaves like the normal one.
