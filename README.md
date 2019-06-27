## SONiC Management Framework Repo

### Build Instruction
Please note that the build instruction in this guide has only been tested on Ubuntu 16.04.
#### Pre-rerequisit

* Packages to be installed:
 * sudo apt-get install git
 * sudo apt-get install docker 

#### Steps to build and create an installer
1. git clone https://github.com/project-arlo/sonic-buildimage.git
2. cd sonic-buildimage/
3. sudo modprobe overlay
4. make init
5. make configure PLATFORM=broadcom
6. Run the prefetch python script to download all binaries (see below for the script).
7. BLDENV=stretch make target/docker-sonic-mgmt-framework.gz 
8. BLDENV=stretch make stretch
9. BLDENV=stretch make target/sonic-broadcom.bin
 
#### Faster builds
In order to speed up the process of build, you can prefetch the latest debian files from Azure server, and just build what you need.

Here is a python script you could use to fetch latest prebuilt objects (deb, gz, ko, etc) from SONiC Jenkins cluster:

    import os
    import shutil
    import urllib.request
    from html.parser import HTMLParser

    UPSTREAM_PREFIX = 'https://sonic-jenkins.westus2.cloudapp.azure.com/job/broadcom/job/buildimage-brcm-all/lastSuccessfulBuild/artifact/'

    def get_all_bins(target_path, extension):
        """Get all files matching the given extension from the target path"""
        print('Fetching %s*%s' % (target_path, extension))
        os.makedirs(target_path, exist_ok=True)

        req = urllib.request.urlopen(UPSTREAM_PREFIX + target_path)
        data = req.read().decode()

        class Downloader(HTMLParser):
            """Class to parse retrieved data, match against the given extension,
               and download the matching files to the given target directory"""
            def handle_starttag(self, tag, attrs):
                """Handle only <a> tags"""
                if tag == 'a':
                    for attr, val in attrs:
                        if attr == 'href' and val.endswith(extension):
                            self.download_file(val)

            @staticmethod
            def download_file(path):
                filename = os.path.join(target_path, path)
                freq = urllib.request.urlopen(UPSTREAM_PREFIX + target_path + path)

                print('\t%s' % path)
                with open(filename, 'wb') as fp:
                    shutil.copyfileobj(freq, fp)


        parser = Downloader()
        parser.feed(data)
        print()

    get_all_bins('target/debs/stretch/', '.deb')
    get_all_bins('target/files/stretch/', '.ko')
    get_all_bins('target/python-debs/', '.deb')
    get_all_bins('target/python-wheels/', '.whl')
    get_all_bins('target/', '.gz')


* Follow steps 1 to 6 in the previous section.
* Download prebuilt objects from azure jenkins server using the above script.
* Then run step 7 to create docker image, and run step 9 to build complete sonic onie NOS installer .


##### Incremental builds 
Just clean up the deb's/gz that require re-build, and build again. Here is an exmple:

##### To build deb file for sonic-mgmt-framework

	BLDENV=stretch make target/debs/stretch/sonic-mgmt-framework_1.0-01_amd64.deb-clean
	BLDENV=stretch make target/debs/stretch/sonic-mgmt-framework_1.0-01_amd64.deb
	
##### To build sonic-mgmt-framework docker alone

	BLDENV=stretch make target/docker-sonic-mgmt-framework.gz-clean
	BLDENV=stretch make target/docker-sonic-mgmt-framework.gz
