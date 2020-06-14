<?xml version="1.0"?>
<!--
Copyright 2019 Dell, Inc.  

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
--> 

<xsl:stylesheet version="1.0"
xmlns:xsl="http://www.w3.org/1999/XSL/Transform">

<!-- Write the selected features of a platform into an xml output file -->
<xsl:output method ="xml" version="1.0" indent="yes"/>
<xsl:template match="/">
 <xsl:element name="root">
  <xsl:for-each select="/PLATFORMMODULE/FEATURELIST/FEATURE">
   <xsl:choose>
    <!-- Case 1 & 2 : (no dynamic or condition attributes set) or dynamic is false -->
    <xsl:when test="not(@dynamic or @condition) or @dynamic='false'">
     <xsl:element name="newfeature">
      <xsl:attribute name="enabled">
        <xsl:value-of select="@enabled"/>
      </xsl:attribute>
      <xsl:value-of select="."/>
     </xsl:element>
    </xsl:when>
    <xsl:otherwise>
     <!-- Case 3 : dynamic is true, condition is available -->
     <xsl:if test="@dynamic='true' and @condition!='' and @namespace!='' and @xpath!='' and @expected_value!=''">
      <xsl:element name="newfeature">
       <xsl:attribute name="enabled">
        <xsl:value-of select="@enabled"/>
       </xsl:attribute>
       <xsl:attribute name="dynamic">
        <xsl:value-of select="@dynamic"/>
       </xsl:attribute>
       <xsl:attribute name="condition">
        <xsl:value-of select="@condition"/>
       </xsl:attribute>
       <xsl:attribute name="namespace">
        <xsl:value-of select="@namespace"/>
       </xsl:attribute>
       <xsl:attribute name="xpath">
        <xsl:value-of select="@xpath"/>
       </xsl:attribute>
       <xsl:attribute name="expected_value">
        <xsl:value-of select="@expected_value"/>
       </xsl:attribute>
       <xsl:value-of select="."/>
      </xsl:element>
     </xsl:if>
    </xsl:otherwise>
   </xsl:choose>
  </xsl:for-each>
 </xsl:element>
</xsl:template>
</xsl:stylesheet>
