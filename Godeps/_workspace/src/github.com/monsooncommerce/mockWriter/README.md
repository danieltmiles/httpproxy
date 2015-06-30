# MockWriter
Library for a mock IO Writer

# Example
```Go
import(
	"fmt"
	"github.com/monsooncommerce/mockwriter"
	"strings"
)

func main(){
	m := mockwriter.New()
	m.Write("A random line to write")

	if strings.Contains( string(m.Written), "random" ){
		fmt.Println("Found a random")
	}
}
```
