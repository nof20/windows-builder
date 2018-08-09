P0
* Build and push a Docker container to GCR
* Build and push one other container, e.g. Go
* Create Windows server automatically

P1

* Move from port 5985 (HTTP) to 5986 (HTTPS)
* Use cloud-build-local.  Looks like it needs a couple of dozen changes.
    - Get benefits from running commands through powershell.exe.  e.g., support for ls, ~, etc.
    - docker.sock the most obvious one.
    - But e.g. hacky use of sed needs thoughtful replacement.

Bind mounting:
* Looks like Docker 17.09.0-ce-rc1 or higher is required - maybe on host as well as client.
* See https://github.com/StefanScherer/insider-docker-machine/pull/1.
* docker run -v '\\.\pipe\docker_engine:\\.\pipe\docker_engine' ...
* To upgrade the Docker daemon, have to do manually.  MS only knows version 17.06.
* https://docs.docker.com/install/windows/docker-ee/#use-a-script-to-install-docker-ee


Revised plan
• Two options: BYO Windows server, or I'll start one for you
• server object has IP, username and password. And embedded GCE object if required.
• GCE object suould have methods attached to it to do stuff. All state should be encoded within itself
• Program flow:
    - Start server if required
    - copy across workspace using WinRMCP
    - Run container build with connection to host Docker sock. Run Local Builder on windows.yaml.
    - Last step should be Docker push. Or, have another env field to support that
    - Copy workspace back.
• Much simpler than previous designs.
• Need to bootstrap a Windows docker container if it doesn't already exist. 

Other notes:
* Powershell container for Linux at microsoft/powershell. e.g. docker run -it microsoft/powershell
* Install VisualStudio build tools into a container:
https://docs.microsoft.com/en-us/visualstudio/install/build-tools-container#step-3-switch-to-windows-containers
* Connect to Docker Windows named pipe:
https://blog.docker.com/2017/09/docker-windows-server-1709/
* Install Jenkins, Git and Docker executable in a container:
https://github.com/jenkinsci/docker/pull/582/files
(super helpful)
* Original issue framing lack of named pipe support: https://github.com/moby/moby/issues/36562
* Official MS docs: https://docs.microsoft.com/en-us/virtualization/windowscontainers/manage-docker/configure-docker-daemon


Providing a startup script.  Powershell script contents:

winrm set winrm/config/Service/Auth @{Basic="true”}
winrm set winrm/config/Service @{AllowUnencrypted="true”}
