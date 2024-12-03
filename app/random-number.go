package app

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strconv"

	m "github.com/sunshine69/webnote-go/models"
)

func GenRandNumber(w http.ResponseWriter, r *http.Request) {
	max_num_str := m.GetRequestValue(r, "mux_num", "999999999999")
	max_num, err := strconv.ParseInt(max_num_str, 10, 64)
	if err != nil {
		max_num = 999999999999
	}
	gen_number, _ := rand.Int(rand.Reader, big.NewInt(max_num))
	fmt.Fprintf(w, "%d", gen_number)
}
