<?xml version="1.0" encoding="UTF-8"?>
<xsl:stylesheet version="1.0" xmlns:xsl="http://www.w3.org/1999/XSL/Transform">
  <xsl:output encoding="UTF-8" method="xml"></xsl:output>

  <xsl:template match="/">
    <testsuite>
      <xsl:attribute name="tests">
        <xsl:value-of select="count(.//file)" />
      </xsl:attribute>
      <xsl:attribute name="failures">
        <xsl:value-of select="count(.//error)" />
      </xsl:attribute>
      <xsl:for-each select="//checkstyle">
        <xsl:apply-templates />
      </xsl:for-each>
    </testsuite>
  </xsl:template>

  <xsl:template match="file">
    <testcase>
      <xsl:attribute name="classname">
        <xsl:value-of select="@name" />
      </xsl:attribute>
      <xsl:attribute name="name">
        <xsl:value-of select="@name" />
      </xsl:attribute>
      <xsl:apply-templates select="node()" />
    </testcase>
  </xsl:template>

  <xsl:template match="error">
    <failure>
      <xsl:attribute name="type">
        <xsl:value-of select="@source" />
      </xsl:attribute>
      <xsl:text>Line </xsl:text>
      <xsl:value-of select="@line" />
      <xsl:text>: </xsl:text>
      <xsl:value-of select="@message" />
      <xsl:text> See https://www.shellcheck.net/wiki/</xsl:text>
      <xsl:value-of select="substring(@source, '12')" />
    </failure>
  </xsl:template>
</xsl:stylesheet>
