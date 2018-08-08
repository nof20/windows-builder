[Net.ServicePointManager]::SecurityProtocol = "tls12, tls11, tls"
Invoke-WebRequest -Uri https://github.com/git-for-windows/git/releases/download/v2.18.0.windows.1/MinGit-2.18.0-64-bit.zip -OutFile git.zip
Expand-Archive -Path git.zip -DestinationPath c:\Git
$env:Path += ";C:\Git\cmd"
git clone https://github.com/nof20/windows-builder.git
# Ignore errors for this, it works.
gcloud auth configure-docker
