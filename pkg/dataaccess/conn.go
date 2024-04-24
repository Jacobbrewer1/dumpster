package dataaccess

import (
	"fmt"

	"github.com/Jacobbrewer1/dumpster/pkg/vault"
	"github.com/spf13/viper"
)

func GenerateConnectionStr(v *viper.Viper, vs vault.Secrets) string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s",
		vs["username"],
		vs["password"],
		v.GetString("db.host"),
		v.GetString("db.schema"),
	)
}
