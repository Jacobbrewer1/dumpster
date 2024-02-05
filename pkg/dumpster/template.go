package dumpster

const tmpl = `
-- Server version	{{ .ServerVersion }}

DROP DATABASE IF EXISTS {{ .Database }};

{{range .Tables}}
-- Table structure for table {{ .Name }}
{{ .SQL }};

-- Data dump for table {{ .Name }}
LOCK TABLES {{ .Name }} WRITE;
{{ if .Values }}
INSERT INTO {{ .Name }} VALUES {{ .Values }};
{{ end }}
UNLOCK TABLES;
{{ end }}

-- Dump completed at {{ .CompleteTime }}
`
