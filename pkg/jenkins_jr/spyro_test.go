package jenkins_jr

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"time"

	"github.com/wiskarindra/jenkins_jr/pkg/currentuser"
	"github.com/wiskarindra/jenkins_jr/pkg/mysql"
	"github.com/wiskarindra/jenkins_jr/pkg/resource"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/subosito/gotenv"
)

var jwtTokens = map[string]string{
	"jwt-token":        "40556304d7dbd92f7fb770162e35fcb9210f6b9614f23917f676d9680af0f1ab",
	"jwt-invalid":      "Token an.invalid.token",
	"jwt-valid":        "Token eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJ0b2tlbiI6IjQwNTU2MzA0ZDdkYmQ5MmY3ZmI3NzAxNjJlMzVmY2I5MjEwZjZiOTYxNGYyMzkxN2Y2NzZkOTY4MGFmMGYxYWIiLCJzY29wZXMiOiJwdWJsaWMgdXNlciBzdG9yZSIsInJlZnJlc2hfdG9rZW4iOiJjNTBjYjc0NmMwMWQ3MmYyODVhMWRmZGUyN2QzNjlhNjIxZDczM2EyM2ZhMDcwODA3OGY5ZTU0OTgxN2JkYzdmIiwicmVzb3VyY2Vfb3duZXJfaWQiOjEsInJlc291cmNlX293bmVyIjp7ImlkIjoxLCJ1c2VybmFtZSI6ImdhbGloIiwibmFtZSI6IkdhbGloIFB1dGVyYSBOdWdyYWhhIFN1bWludG8iLCJvZmZpY2lhbCI6ZmFsc2UsInZlcmlmaWVkIjp0cnVlLCJlbWFpbCI6ImdhbGloLnB1dGVyYTk0QGdtYWlsLmNvbSIsImdlbmRlciI6Ikxha2ktbGFraSIsInBob25lIjoiMDg1MjEwMTM1NjEiLCJyb2xlIjoiYWRtaW4iLCJsYXN0X2xvZ2luX2F0IjoiMjAxNy0wNi0xMlQyMDoyNTo0MCswNzowMCIsImpvaW5lZF9hdCI6IjIwMTUtMDgtMzFUMTQ6NTI6MTIrMDc6MDAiLCJwcmVtaXVtX3N1YnNjcmlwdGlvbl9sZXZlbCI6InByb2Zlc3Npb25hbCJ9LCJhcHBsaWNhdGlvbl9pZCI6MSwiYXBwbGljYXRpb25fbmFtZSI6ImNsaWVudF9hcHBzIiwiYXBwbGljYXRpb25fc2NvcGVzIjoicHVibGljIHVzZXIgc3RvcmUiLCJhcHBsaWNhdGlvbl91aWQiOiIwNGUzYjQ1MTc3MjIwNWFmZDdmNmExMzUifQ.bvjlVacTPfV_jIoV1PI1_9-Z_kaN6bpf7Gz89yijXck3KxGC-IbKF_GdFgJvHaeF9gwQtM39Nno1iFtpmHy3vSFKifc0-b9WBiTAlC3KOKmF0XBsv9L1guwnmtA108eoDETPazKG_DaFWx1BWSRNd-Xw-z8vKlVJhgoL0j3j1gUQcKjMycuozWyPLEQHAbvAazYylAmj8FvfjeStVfpLtlfLaytlOgiq4p5J93Wkw_2EOegIO6lOGFJYsVoqHoPOgSfWKS6pJjanOfgwOBL-RtPVMfB-1hkNH50ev9xLZ4visD9Obv281LTGovcvU69dJUsFdgTdfJ8g5GmC8-P6DA",
	"jwt-valid-female": "Token eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJ0b2tlbiI6IjQwNTU2MzA0ZDdkYmQ5MmY3ZmI3NzAxNjJlMzVmY2I5MjEwZjZiOTYxNGYyMzkxN2Y2NzZkOTY4MGFmMGYxYWIiLCJzY29wZXMiOiJwdWJsaWMgdXNlciBzdG9yZSIsInJlZnJlc2hfdG9rZW4iOiJjNTBjYjc0NmMwMWQ3MmYyODVhMWRmZGUyN2QzNjlhNjIxZDczM2EyM2ZhMDcwODA3OGY5ZTU0OTgxN2JkYzdmIiwicmVzb3VyY2Vfb3duZXJfaWQiOjEsInJlc291cmNlX293bmVyIjp7ImlkIjoxLCJ1c2VybmFtZSI6ImdhbGloIiwibmFtZSI6IkdhbGloIFB1dGVyYSBOdWdyYWhhIFN1bWludG8iLCJvZmZpY2lhbCI6ZmFsc2UsInZlcmlmaWVkIjp0cnVlLCJlbWFpbCI6ImdhbGloLnB1dGVyYTk0QGdtYWlsLmNvbSIsImdlbmRlciI6IlBlcmVtcHVhbiIsInBob25lIjoiMDg1MjEwMTM1NjEiLCJyb2xlIjoiYWRtaW4iLCJsYXN0X2xvZ2luX2F0IjoiMjAxNy0wNi0xMlQyMDoyNTo0MCswNzowMCIsImpvaW5lZF9hdCI6IjIwMTUtMDgtMzFUMTQ6NTI6MTIrMDc6MDAiLCJwcmVtaXVtX3N1YnNjcmlwdGlvbl9sZXZlbCI6InByb2Zlc3Npb25hbCJ9LCJhcHBsaWNhdGlvbl9pZCI6MSwiYXBwbGljYXRpb25fbmFtZSI6ImNsaWVudF9hcHBzIiwiYXBwbGljYXRpb25fc2NvcGVzIjoicHVibGljIHVzZXIgc3RvcmUiLCJhcHBsaWNhdGlvbl91aWQiOiIwNGUzYjQ1MTc3MjIwNWFmZDdmNmExMzUifQ.Eq3bi1ZG1KEWftGsvqeCCMswVVfkq9BUbOQ6E4lVE0qFCjyeUtQNOmFPiuSXl6aS1hfxocqQD6JFK4CwpW1AVm21yPvWMhMHNVAkKhLLNm-gm2q_OdRBScBYyEEN8_fcgKuTfLrh33F7fmbLKf2uXN4Th_iJpJ8mFBmfVh3J-ad6U056KhfFUnRq_dY-AIX8whh8cl47az4xIbNwT7Uui5VHNVAH7JfmFyRPr5XelPdcUpBC8YC3p19UD0OUfaFc6Yv0wDWUYqeQRow_y2Kpiu4r1PmxymwxqhpUy6MfxiqzA5KjBqIr0VI9WN8WudgJZnT-6Wn3co-dnFZS9b25LA",
	"jwt-valid-logout": "Token eyJhbGciOiJSUzI1NiIsImtpZCI6IiJ9.eyJ0b2tlbiI6IjQwNTU2MzA0ZDdkYmQ5MmY3ZmI3NzAxNjJlMzVmY2I5MjEwZjZiOTYxNGYyMzkxN2Y2NzZkOTY4MGFmMGYxYWIiLCJzY29wZXMiOiJwdWJsaWMgc3RvcmUiLCJyZWZyZXNoX3Rva2VuIjoiYzUwY2I3NDZjMDFkNzJmMjg1YTFkZmRlMjdkMzY5YTYyMWQ3MzNhMjNmYTA3MDgwNzhmOWU1NDk4MTdiZGM3ZiIsImFwcGxpY2F0aW9uX2lkIjoxLCJhcHBsaWNhdGlvbl9uYW1lIjoiY2xpZW50X2FwcHMiLCJhcHBsaWNhdGlvbl9zY29wZXMiOiJwdWJsaWMgdXNlciBzdG9yZSIsImFwcGxpY2F0aW9uX3VpZCI6IjA0ZTNiNDUxNzcyMjA1YWZkN2Y2YTEzNSJ9.y-qTSaU-b8b-ZZW3ZJMOQ1hT4VC_Zk1O-_TQPR6eD2wuKF1lHXtLL-TDtoGhBcaF-R4GPIQR6R3aOfnFaqIUgsGA3udqLeeQM-UfXbTm36_IcCjdGivFXDUVeYQ6f3zUNT8N0Y8BPhZm-4JnY2lHB_Gb1Wkk2-LtF1dEvgFyhEV7nT-XFxrFLDmLwZ-sNR0it-XNbO66xnFsFHImlO1HjLT_3DL898RL304jJ8p4QMx5GeqGliGVGCIlW5LdLBqjQmHJtUhbLhk9tXDCC_GPE0hT2ac19bTwW1BV7i9ahhkZrZ3IO-hQbINQQAIsR_zjbgnoMGohAcn5cQentdPCzw",
}

func init() {
	gotenv.MustLoad(os.Getenv("GOPATH") + "/src/github.com/wiskarindra/jenkins_jr/.env")
	os.Setenv("ENV", "test")
}

func newEnvTest() Env {
	db := initMySQLTest()
	return Env{DB: db}
}

func initMySQLTest() *sqlx.DB {
	os.Setenv("DATABASE_NAME", os.Getenv("DATABASE_TEST_NAME"))
	os.Setenv("DATABASE_PORT", os.Getenv("DATABASE_TEST_PORT"))
	os.Setenv("DATABASE_USERNAME", os.Getenv("DATABASE_TEST_USERNAME"))
	os.Setenv("DATABASE_PASSWORD", os.Getenv("DATABASE_TEST_PASSWORD"))
	os.Setenv("DATABASE_HOST", os.Getenv("DATABASE_TEST_HOST"))
	return mysql.Init()
}

func clearDBTestData(db *sqlx.DB) {
	tables := []string{
		"action_log_histories",
		"categories",
		"post_filters",
		"post_images",
		"post_likes",
		"post_tags",
		"posts",
	}

	for _, table := range tables {
		db.Exec("DELETE FROM " + table)
	}
}

func newRequestTest(body []byte, token string) (*httptest.ResponseRecorder, *http.Request) {
	startTime := time.Now()

	r := httptest.NewRequest("GET", "http://example.com/foo", strings.NewReader(string(body)))
	r.Header.Set("Authorization", jwtTokens[token])
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	currentUser, _ := currentuser.FromRequest(r)
	ctx := currentuser.NewContext(r.Context(), currentUser)
	ctx = resource.NewContext(ctx, "0", "call-just-for-testing", startTime)

	r = r.WithContext(ctx)
	return w, r
}
