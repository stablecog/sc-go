package shared

import (
	"context"
	"strings"

	"github.com/stablecog/sc-go/database/ent"
	"github.com/stablecog/sc-go/database/ent/disposableemail"
	"github.com/stablecog/sc-go/log"
)

func IsDisposableEmail(email string, db *ent.Client, ctx context.Context) bool {
	if !strings.Contains(email, "@") {
		exists, err := db.DisposableEmail.Query().Where(disposableemail.Domain(email)).Exist(ctx)
		if err != nil {
			log.Errorf("🔴 📨 Error checking if email exists: %v", err)
			return false
		}
		if exists {
			log.Infof("🔴 📨 Disposable domain without @ symbol - Email: %s - Disposable Domain: %s", email, email)
			return true
		}
		return false
	}

	segs := strings.Split(email, "@")
	if len(segs) != 2 {
		return false
	}
	domain := strings.ToLower(segs[1])
	exists, err := db.DisposableEmail.Query().Where(disposableemail.Domain(email)).Exist(ctx)
	if err != nil {
		log.Errorf("🔴 📨 Error checking if email exists: %v", err)
		return false
	}
	if exists {
		log.Infof("🔴 📨 Disposable domain - Email: %s - Domain: %s", email, domain)
		return true
	}
	return false
}
