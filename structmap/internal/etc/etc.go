package etc
	
import (
	"math/rand"
	"time"
)

var Crand = rand.NewSource(time.Now().Unix())
