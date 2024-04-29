package dumpster

const tmpl = `
-- Server version	{{ .ServerVersion }}

CREATE DATABASE IF NOT EXISTS {{ .Database }};
USE {{ .Database }};

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

{{ range .Triggers }}
-- Trigger structure for trigger {{ .Name }}
{{ .SQL }};
{{ end }}

-- Dump completed at {{ .CompleteTime }}
`
