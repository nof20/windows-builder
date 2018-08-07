# Design

High-level design is as follows:

* Customers provide a Windows build machine, or we start one on GCE.
* GCB workspace is synchronized to `C:\workspace` on the Windows machine.
* Customers run build steps in containers, as they do at present on GCB.
* Workspace is synced back.

Windows build steps are provided in a separate YAML file.  This avoids the complexity and potential confusion of having to provide a series of arguments in the original Cloud Build YAML file.

A Docker on Windows build step must also be provided.  

Compatibility:

* 