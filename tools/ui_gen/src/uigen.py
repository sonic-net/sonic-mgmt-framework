#! /usr/bin/env python3
import os
import glob
from jinja2 import Environment, FileSystemLoader

currentDir = os.path.dirname(os.path.realpath(__file__))
templateEnv = Environment(loader=FileSystemLoader(os.path.join(currentDir,'../templates')),trim_blocks=True,lstrip_blocks=True)
yamlDir = os.path.join(currentDir, '../../../build/rest_server/dist/ui')
yamlFilesList = glob.glob(os.path.join(yamlDir,'*.yaml'))
yamlList = sorted(list(map(lambda x: os.path.splitext(os.path.basename(x))[0], yamlFilesList)))

#Generate Swagger index.html file
swaggerIndexHtmlStr = templateEnv.\
	get_template('swagger_index_html.template').\
	render(yamlList=yamlList)

#Generate landing page landing.html file
landingHtmlStr = templateEnv.\
	get_template('landing_html.template').\
	render(yamlList=yamlList)

swaggerIndexHtmlFilePath = os.path.join(currentDir, '../../../build/rest_server/dist/ui/model.html')
landingHtmlFilePath = os.path.join(currentDir, '../../../build/rest_server/dist/ui/index.html')

#write swagger index html file
with open(swaggerIndexHtmlFilePath, "w") as swaggerIndexFh:
	swaggerIndexFh.write(swaggerIndexHtmlStr)

#write landing html file
with open(landingHtmlFilePath, "w") as landingFh:
	landingFh.write(landingHtmlStr)
