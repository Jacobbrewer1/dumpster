package dumpster

const tmpl = `
-- Server version	{{ .ServerVersion }}

{{range .Tables}}
-- Table structure for table {{ .Name }}
DROP TABLE IF EXISTS {{ .Name }};
{{ .SQL }};

-- Data dump for table {{ .Name }}
LOCK TABLES {{ .Name }} WRITE;
/*!40000 ALTER TABLE {{ .Name }} DISABLE KEYS */;
{{ if .Values }}
INSERT INTO {{ .Name }} VALUES {{ .Values }};
{{ end }}
/*!40000 ALTER TABLE {{ .Name }} ENABLE KEYS */;
UNLOCK TABLES;
{{ end }}

-- Dump completed on {{ .CompleteTime }}
`
