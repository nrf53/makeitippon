module github.com/nrf53/makeitippon

go 1.22.3

//replace github.com/nrf53/makeitippon => /home/nrf53/makeitippon_improvement/makeitippon/sushi
replace github.com/nrf53/makeitippon/internal/pkg/sushi => /home/nrf53/makeitippon_improvement/makeitippon/internal/pkg/sushi

replace github.com/nrf53/makeitippon/internal/pkg/image => /home/nrf53/makeitippon_improvement/makeitippon/internal/pkg/image

require github.com/nrf53/makeitippon/internal/pkg/sushi v0.0.0-00010101000000-000000000000

require (
	github.com/nrf53/makeitippon/internal/pkg/image v0.0.0-00010101000000-000000000000 // indirect
	golang.org/x/image v0.16.0 // indirect
	golang.org/x/net v0.25.0 // indirect
	golang.org/x/text v0.15.0 // indirect
)
