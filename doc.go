/**
Data source name

scheme://username:password@host:port/dbname?param1=value1&param2=value2&...

d, err := dsn.Parse("scheme://username:password@host:port/dbname?param1=value1&param2=value2")

var cfg struct{
	Address string `dsn:"address"`
}
err := d.Bind(&cfg)

*/
package gdsn
