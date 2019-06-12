<?xml version="1.0"?>
<xsl:stylesheet version="1.0"
xmlns:xsl="http://www.w3.org/1999/XSL/Transform">

<!-- Print the ENTITY definition string, looking at the input XML file -->
<xsl:output method ="text"/>
<xsl:template match="/">
  <xsl:for-each select="/PLATFORMMODULE/ENTITYLIST/ENTITYNAME">
&lt;!ENTITY <xsl:value-of select="."/> "<xsl:value-of select="./@value"/>"&gt;</xsl:for-each>
</xsl:template>
</xsl:stylesheet>
