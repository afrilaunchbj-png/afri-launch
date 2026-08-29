// Package pptx assemble un fichier PowerPoint minimal où chaque slide est
// une image pleine page (image-par-slide, cf. ADR-013).
package pptx

import (
	"archive/zip"
	"bytes"
	"fmt"
	"strings"
)

const (
	nsA   = "http://schemas.openxmlformats.org/drawingml/2006/main"
	nsP   = "http://schemas.openxmlformats.org/presentationml/2006/main"
	nsR   = "http://schemas.openxmlformats.org/officeDocument/2006/relationships"
	nsRel = "http://schemas.openxmlformats.org/package/2006/relationships"
)

const contentTypesXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">
<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/>
<Default Extension="xml" ContentType="application/xml"/>
<Default Extension="png" ContentType="image/png"/>
<Override PartName="/ppt/presentation.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.presentation.main+xml"/>
<Override PartName="/ppt/slideMasters/slideMaster1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideMaster+xml"/>
<Override PartName="/ppt/slideLayouts/slideLayout1.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slideLayout+xml"/>
<Override PartName="/ppt/theme/theme1.xml" ContentType="application/vnd.openxmlformats-officedocument.theme+xml"/>
%s</Types>`

const rootRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="` + nsRel + `">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="ppt/presentation.xml"/>
</Relationships>`

const presentationXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:presentation xmlns:a="` + nsA + `" xmlns:r="` + nsR + `" xmlns:p="` + nsP + `">
<p:sldMasterIdLst><p:sldMasterId id="2147483648" r:id="rId1"/></p:sldMasterIdLst>
<p:sldIdLst>%s</p:sldIdLst>
<p:sldSz cx="%d" cy="%d" type="screen16x9"/>
<p:notesSz cx="6858000" cy="9144000"/>
</p:presentation>`

const presentationRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="` + nsRel + `">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster" Target="slideMasters/slideMaster1.xml"/>
%s</Relationships>`

const slideMasterXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sldMaster xmlns:a="` + nsA + `" xmlns:r="` + nsR + `" xmlns:p="` + nsP + `">
<p:cSld><p:spTree><p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr><p:grpSpPr/></p:spTree></p:cSld>
<p:clrMap bg1="lt1" tx1="dk1" bg2="lt2" tx2="dk2" accent1="accent1" accent2="accent2" accent3="accent3" accent4="accent4" accent5="accent5" accent6="accent6" hlink="hlink" folHlink="folHlink"/>
<p:sldLayoutIdLst><p:sldLayoutId id="2147483649" r:id="rId1"/></p:sldLayoutIdLst>
</p:sldMaster>`

const slideMasterRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="` + nsRel + `">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout" Target="../slideLayouts/slideLayout1.xml"/>
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/theme" Target="../theme/theme1.xml"/>
</Relationships>`

const slideLayoutXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sldLayout xmlns:a="` + nsA + `" xmlns:r="` + nsR + `" xmlns:p="` + nsP + `" type="blank" preserve="1">
<p:cSld name="Blank"><p:spTree><p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr><p:grpSpPr/></p:spTree></p:cSld>
<p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr>
</p:sldLayout>`

const slideLayoutRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="` + nsRel + `">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideMaster" Target="../slideMasters/slideMaster1.xml"/>
</Relationships>`

const slideXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<p:sld xmlns:a="` + nsA + `" xmlns:r="` + nsR + `" xmlns:p="` + nsP + `">
<p:cSld><p:spTree>
<p:nvGrpSpPr><p:cNvPr id="1" name=""/><p:cNvGrpSpPr/><p:nvPr/></p:nvGrpSpPr>
<p:grpSpPr/>
<p:pic>
<p:nvPicPr><p:cNvPr id="2" name="Image"/><p:cNvPicPr><a:picLocks noChangeAspect="1"/></p:cNvPicPr><p:nvPr/></p:nvPicPr>
<p:blipFill><a:blip r:embed="rId2"/><a:stretch><a:fillRect/></a:stretch></p:blipFill>
<p:spPr><a:xfrm><a:off x="0" y="0"/><a:ext cx="%d" cy="%d"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></p:spPr>
</p:pic>
</p:spTree></p:cSld>
<p:clrMapOvr><a:masterClrMapping/></p:clrMapOvr>
</p:sld>`

const slideRelsXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<Relationships xmlns="` + nsRel + `">
<Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slideLayout" Target="../slideLayouts/slideLayout1.xml"/>
<Relationship Id="rId2" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="../media/image%d.png"/>
</Relationships>`

const themeXML = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<a:theme xmlns:a="` + nsA + `" name="AfriLaunch">
<a:themeElements>
<a:clrScheme name="Emerald Amber">
<a:dk1><a:sysClr val="windowText" lastClr="000000"/></a:dk1>
<a:lt1><a:sysClr val="window" lastClr="FFFFFF"/></a:lt1>
<a:dk2><a:srgbClr val="191C1D"/></a:dk2>
<a:lt2><a:srgbClr val="F8F9FA"/></a:lt2>
<a:accent1><a:srgbClr val="003527"/></a:accent1>
<a:accent2><a:srgbClr val="855300"/></a:accent2>
<a:accent3><a:srgbClr val="059669"/></a:accent3>
<a:accent4><a:srgbClr val="4F1F19"/></a:accent4>
<a:accent5><a:srgbClr val="FEA619"/></a:accent5>
<a:accent6><a:srgbClr val="707974"/></a:accent6>
<a:hlink><a:srgbClr val="003527"/></a:hlink>
<a:folHlink><a:srgbClr val="855300"/></a:folHlink>
</a:clrScheme>
<a:fontScheme name="AfriLaunch">
<a:majorFont><a:latin typeface="Lexend"/></a:majorFont>
<a:minorFont><a:latin typeface="Inter"/></a:minorFont>
</a:fontScheme>
<a:fmtScheme name="AfriLaunch">
<a:fillStyleLst><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:fillStyleLst>
<a:lnStyleLst><a:ln w="9525" cap="flat" cmpd="sng" algn="ctr"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:prstDash val="solid"/></a:ln><a:ln w="25400" cap="flat" cmpd="sng" algn="ctr"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:prstDash val="solid"/></a:ln><a:ln w="38100" cap="flat" cmpd="sng" algn="ctr"><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:prstDash val="solid"/></a:ln></a:lnStyleLst>
<a:effectStyleLst><a:effectStyle><a:effectLst/></a:effectStyle><a:effectStyle><a:effectLst/></a:effectStyle><a:effectStyle><a:effectLst/></a:effectStyle></a:effectStyleLst>
<a:bgFillStyleLst><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:solidFill><a:schemeClr val="phClr"/></a:solidFill><a:solidFill><a:schemeClr val="phClr"/></a:solidFill></a:bgFillStyleLst>
</a:fmtScheme>
</a:themeElements>
</a:theme>`

// Build assemble un PPTX (image-par-slide) à partir d'images PNG.
func Build(images [][]byte, widthEMU, heightEMU int) ([]byte, error) {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	write := func(name, content string) error {
		w, err := zw.Create(name)
		if err != nil {
			return err
		}
		_, err = w.Write([]byte(content))
		return err
	}

	// Slide overrides + rels + ids.
	var slideOverrides, slideIDs, slideRels strings.Builder
	for i := range images {
		n := i + 1
		slideOverrides.WriteString(fmt.Sprintf(
			`<Override PartName="/ppt/slides/slide%d.xml" ContentType="application/vnd.openxmlformats-officedocument.presentationml.slide+xml"/>`+"\n", n))
		slideIDs.WriteString(fmt.Sprintf(`<p:sldId id="%d" r:id="rId%d"/>`, 255+n, 1+n))
		slideRels.WriteString(fmt.Sprintf(
			`<Relationship Id="rId%d" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/slide" Target="slides/slide%d.xml"/>`+"\n", 1+n, n))
	}

	parts := []struct{ name, content string }{
		{"[Content_Types].xml", fmt.Sprintf(contentTypesXML, slideOverrides.String())},
		{"_rels/.rels", rootRelsXML},
		{"ppt/presentation.xml", fmt.Sprintf(presentationXML, slideIDs.String(), widthEMU, heightEMU)},
		{"ppt/_rels/presentation.xml.rels", fmt.Sprintf(presentationRelsXML, slideRels.String())},
		{"ppt/slideMasters/slideMaster1.xml", slideMasterXML},
		{"ppt/slideMasters/_rels/slideMaster1.xml.rels", slideMasterRelsXML},
		{"ppt/slideLayouts/slideLayout1.xml", slideLayoutXML},
		{"ppt/slideLayouts/_rels/slideLayout1.xml.rels", slideLayoutRelsXML},
		{"ppt/theme/theme1.xml", themeXML},
	}
	for _, p := range parts {
		if err := write(p.name, p.content); err != nil {
			return nil, err
		}
	}

	for i, img := range images {
		n := i + 1
		if err := write(fmt.Sprintf("ppt/slides/slide%d.xml", n), fmt.Sprintf(slideXML, widthEMU, heightEMU)); err != nil {
			return nil, err
		}
		if err := write(fmt.Sprintf("ppt/slides/_rels/slide%d.xml.rels", n), fmt.Sprintf(slideRelsXML, n)); err != nil {
			return nil, err
		}
		w, err := zw.Create(fmt.Sprintf("ppt/media/image%d.png", n))
		if err != nil {
			return nil, err
		}
		if _, err := w.Write(img); err != nil {
			return nil, err
		}
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
