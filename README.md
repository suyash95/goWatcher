# goWatcher

It is basic go development server, which detects the changes done in the code directory and auto generate the build to run. In case build gets failed, it will run the previous successful build.

# How to run

``` go run gowatcher.go "env varibale" "absolute path to main.go file" "path to directory to be watched" ```

