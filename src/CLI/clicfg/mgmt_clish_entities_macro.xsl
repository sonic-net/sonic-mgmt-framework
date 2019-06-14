<?xml version="1.0"?>
<xsl:stylesheet version="1.0"
xmlns:xsl="http://www.w3.org/1999/XSL/Transform">

<!-- Write the selected entities of a platform into an xml output file -->
<xsl:output method ="xml" version="1.0" indent="yes"/>
<xsl:template match="/">
 <xsl:element name="root">
  <xsl:for-each select="/PLATFORMMODULE/ENTITYLIST/ENTITYNAME">
   <xsl:element name="{.}">
    <xsl:value-of select="./@value"/>
   </xsl:element>
  </xsl:for-each>
 </xsl:element>
</xsl:template>
</xsl:stylesheet>
