    $rootDir = Join-Path -Path $PSScriptRoot -ChildPath ".."
    Set-Location $rootDir
    $env:GOOS="linux";
    $env:GOARCH="arm";
    $env:GOARM=7;
    $env:CGO_ENABLED="0";
    go build -o _bin/rom-downloader