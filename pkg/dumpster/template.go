package dumpster

const tmpl = `
-- Server version	{{ .ServerVersion }}

DROP DATABASE IF EXISTS {{ .Database }};

SET FOREIGN_KEY_CHECKS=0;
{{ range .Tables}}
-- Table structure for table {{ .Name }}
{{ .SQL }};
{{ if .Values }}
-- Data dump for table {{ .Name }}
LOCK TABLES {{ .Name }} WRITE;

INSERT INTO {{ .Name }} VALUES {{ .Values }};

UNLOCK TABLES;
{{ end }}
{{- end }}

SET FOREIGN_KEY_CHECKS=1;

-- Dump completed at {{ .CompleteTime }}
`
