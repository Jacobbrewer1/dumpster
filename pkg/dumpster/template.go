package dumpster

const tmpl = `
-- Server version	{{ .ServerVersion }}

DROP DATABASE IF EXISTS {{ .Database }};

{{range .Tables}}
-- Table structure for table {{ .Name }}
{{ .SQL }};

{{ if .Values }}
-- Data dump for table {{ .Name }}
LOCK TABLES {{ .Name }} WRITE;

INSERT INTO {{ .Name }} VALUES {{ .Values }};

UNLOCK TABLES;
{{ end }}
{{ end }}

-- Dump completed at {{ .CompleteTime }}
`
