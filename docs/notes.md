P0
* Build and push a Docker container to GCR
* Build and push one other container, e.g. Go
* Create Windows server automatically

P1

Bind mounting:
* Looks like Docker 17.09.0-ce-rc1 or higher is required - maybe on host as well as client.
* See https://github.com/StefanScherer/insider-docker-machine/pull/1.
* docker run -v '\\.\pipe\docker_engine:\\.\pipe\docker_engine' ...
* To upgrade the Docker daemon:

# Check version
Get-Package -Name Docker -ProviderName DockerMsftProvider
# Find out what's available
Find-Package -Name Docker -ProviderName DockerMsftProvider
# Upgrade
Install-Package -Name Docker -ProviderName DockerMsftProvider -Update -Force, followed by Start-Service Docker

* Fallback plan: https://docs.microsoft.com/en-us/virtualization/windowscontainers/management/manage_remotehost
* i.e., use dockertls to create certificates, then listen on TCP socket.  But requires changes to Dockerfile inside docker-windows container, which is gross.

Revised plan
• Two options: BYO Windows server, or I'll start one for you
• server object has IP, username and password. And embedded GCE object if required.
• GCE object suould have methods attached to it to do stuff. All state should be encoded within itself
• Program flow:
    - Start server if required
    - copy across workspace using WinRMCP
    - Run container build with connection to host Docker sock over TCP. Mount local workspace in C:\workspace.  Maybe start with this, prove it works first. This is the core of the problem.
    - Last step should be Docker push. Or, have another env field to support that
    - copy workspace back
• Much simpler than previous attempts
• Need to bootstrap a Windows docker container if it doesn't already exist. 

Resources:
• https://codelabs.developers.google.com/codelabs/cloud-windows-containers-computeengine/#2
• https://cloudplatform.googleblog.com/2018/04/how-to-run-Windows-Containers-on-Compute-Engine.html
• https://m-12.net/copy-files-windows-linux-using-powershell-remoting/
• https://www.thomasmaurer.ch/2017/11/install-ssh-on-windows-10-as-optional-feature/
• https://docs.microsoft.com/en-us/powershell/scripting/core-powershell/running-remote-commands?view=powershell-6
* Powershell container for Linux at microsoft/powershell. e.g. docker run -it microsoft/powershell

• Connect using Powershell:

$passwd = convertto-securestring -AsPlainText -Force -String "@&u+=62qLM^z|g:"   
$cred = new-object -typename System.Management.Automation.PSCredential -argumentlist "franklinn",$passwd   
Enter-PSSession -ComputerName 35.184.87.86 -Credential $cred  

Install VisualStudio build tools into a container:
https://docs.microsoft.com/en-us/visualstudio/install/build-tools-container#step-3-switch-to-windows-containers

Connect to Docker Windows named pipe:
https://blog.docker.com/2017/09/docker-windows-server-1709/

Install Jenkins, Git and Docker executable in a container:
https://github.com/jenkinsci/docker/pull/582/files
(super helpful)

https://github.com/moby/moby/issues/36562

In CMD.EXE:
sc config docker binpath= "\"C:\Program Files\docker\dockerd.exe\" --run-service -H tcp://0.0.0.0:2375"
^^ enables unauthenticated access to Docker daemon on port 2375

Docs say you can provide multiple -H.  https://docs.docker.com/engine/reference/commandline/dockerd/#daemon-socket-option

Looks like the pipe is actually npipe:////./pipe/docker_engine - not \\.\... as the docs say.

In Powershell:
New-NetFirewallRule -DisplayName 'Docker' -Profile @('Domain', 'Private', 'Public') -Direction Inbound -Action Allow -Protocol TCP -LocalPort @('2375')
Restart-Service docker

May be possible to restrict further away from Public.

Then, on command line:
docker -H winvm2:2375 images

...shows we can connect over TCP.  If you build this into the ENTRYPOINT inside the docker container, it can indeed run and connect to the host daemon.  For example:

docker -H localhost:2375 run -it docker-windows images

See e.g.: https://docs.microsoft.com/en-us/virtualization/windowscontainers/manage-docker/configure-docker-daemon

https://dille.name/blog/2017/11/29/using-the-docker-named-pipe-as-a-non-admin-for-windowscontainers/

Rethinking the design
It’s kind-of important that we allow people to build and push Windows containers - that’s the way everything is going
So still use the Windows image with Docker pre-installed, but just run a script (provided in the build image) instead - don’t try and run a container
Script could say “docker build …” or just run whatever else is in the workspace.
Don't mess with external IPs: instead, use instance.c.project.internal. But I wonder if the GCB network will allow that?

Create firewall rule allowing ingress on port 5986
Create the Windows VM with Server DC Core for Containers Image, and provide a startup.bat which allows basic authentication
Reset the Windows password using gcloud (or Python API) and store
Use PyWinRM (or the Go equivalent) to connect to the instance and run commands
gcloud comes ready-installed (thank god), so can use that to sync files to/from GCS
Probably want people to run their builds inside a container on WIndows, maybe need to run gcloud auth configure-docker?  But I wonder if that works on Windows?

gcloud compute --project=nofranklin-test firewall-rules create allow-powershell --direction=INGRESS --priority=1000 --network=default --action=ALLOW --rules=tcp:5986,udp:5986 --source-ranges=0.0.0.0/0

Connecting from desktop Windows Powershell session

From Windows Powershell session (works):

Enter-PSSession -ComputerName 35.226.180.135 -UseSSL -SessionOption (New-PSSessionOption -SkipCACheck -SkipCNCheck) -Credential $cred

Doesn’t work from Powershell container on Linux.  Don’t know why, documentaton is full of acronyms I don’t understand.

Connecting from PyWinRM library

(Works)

from winrm.protocol import Protocol
p = Protocol(
    endpoint='https://35.226.180.135:5986/wsman',
    transport='ntlm',
    username=r'n_o_franklin',
    password='...',
    server_cert_validation='ignore')
shell_id = p.open_shell()
command_id = p.run_command(shell_id, 'ipconfig', ['/all'])
std_out, std_err, status_code = p.get_command_output(shell_id, command_id)
print std_out

Can run Docker containers - it works.

Connecting from golang WinRM library

https://github.com/masterzen/winrm

Enable basic auth and remote connections, etc. by providing a gcloud startup script

Providing a startup script.  Script (.bat) contents:

winrm set winrm/config/Service/Auth @{Basic="true”}
winrm set winrm/config/Service @{AllowUnencrypted="true”}

Installing Container Builder (Cloud Build) Local on Windows

To edit files on Windows: notepad X
To start a new shell: start cmd.exe
Remove a dir: rmdir /s c:\…
Also, seriously - consider using a beefier VM to do Windows builds.  n1-standard-1 just isn’t up to it.

gcloud components install cloud-build-local doesn’t work.  Can now:

gsutil cp gs://nofranklin-windows-builder/cloud-build-local.exe . # 13 MB

Or, bootstrap by building from source.  Download Go:

powershell.exe
[Net.ServicePointManager]::SecurityProtocol = "tls12, tls11, tls"
$client = new-object System.Net.WebClient
$client.DownloadFile("https://dl.google.com/go/go1.10.3.windows-386.msi”,”install.msi")

Somehow install quietly (don’t know how) - just running on command line requires acceptance of dialog boxes, etc.  Looks like /quiet does this? - but beware, it returns immediately and continues in the background, so would have to poll until it completes (or completes very quickly)

set PATH=%PATH%;C:\Go\bin
go version # should show version

$client.DownloadFile("https://github.com/git-for-windows/git/releases/download/v2.18.0.windows.1/Git-2.18.0-32-bit.exe", "git.exe”)

Run git.exe silently (somehow) and install git in C:\Git
More complicated than Go: see e.g. this.  Maybe just /SILENT ?
set PATH=%PATH%;C:\Git\bin

git clone https://github.com/GoogleCloudPlatform/cloud-build-local c:\Go\src\github.com\GoogleCloudPlatform\cloud-build-local
go install github.com/GoogleCloudPlatform/cloud-build-local

Minimal Windows container:

FROM microsoft/windowsservercore:1709
CMD ["hostname”]

Note that the OS version must match the Windows server version, otherwise Docker barfs.  Much less forgiving than Linux.

Minimal cloudbuild.yaml:

steps:
- name: 'gcr.io/cloud-builders/docker'
  args: [ 'build', '-t', 'gcr.io/$PROJECT_ID/quickstart-image', '.' ]
images:
- 'gcr.io/$PROJECT_ID/quickstart-image’

I think I have a bootstrap problem here: the Docker image typically used is a Linux one, so we need a Windows Docker (executable)-in-Docker (platform) container.  cloud-build-local seems to know that, and runs with a -privileged flag.  So let’s try creating one using this hint and see if that works.

gcloud auth configure-docker doesn’t work properly on Windows: have to edit C:\Users\n_o_franklin\.docker\config.json to replace “gcloud” with “gcloud.cmd” to make it work.  See this hint.

Nope - this isn’t going to work.  cloud-build-local prints the Docker command line it runs, for example:

docker run --name step_0 --volume /var/run/docker.sock:/var/run/docker.sock --volume cloudbuild_vol_92a04e80-8de0-4265-9713-329084ccc72f:/workspace --workdir /workspace --volume homevol:/builder/home --env HOME=/builder/home --network cloudbuild --privileged gcr.io/nofranklin-test/docker-windows build -t gcr.io/nofranklin-test/test .

This relies on being able to connect to the host Docker daemon (using the docker.sock volume mount).  It isn’t in /var/run on Windows, so we’d have to tweak that, and perhaps other stuff too.  I’m not into that right now.



