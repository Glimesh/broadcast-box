go build -o BroadcastBox.exe
if ($LASTEXITCODE -eq 0) {
  .\BroadcastBox.exe $args
}
