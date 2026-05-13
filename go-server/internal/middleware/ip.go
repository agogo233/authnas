package middleware

import (
	"net"
	"strings"

	"github.com/authnas/authnas/go-server/internal/config"
	"github.com/gin-gonic/gin"
)

var trustedProxyIPs []net.IP

func InitTrustedProxies(cfg *config.Config) {
	trustedProxyIPs = nil
	for _, proxy := range cfg.Server.TrustedProxies {
		if ip := net.ParseIP(proxy); ip != nil {
			trustedProxyIPs = append(trustedProxyIPs, ip)
		} else {
			_, ipNet, err := net.ParseCIDR(proxy)
			if err == nil {
				trustedProxyIPs = append(trustedProxyIPs, ipNet.IP)
			}
		}
	}
}

func isTrustedProxy(ip net.IP) bool {
	for _, trusted := range trustedProxyIPs {
		if ip.Equal(trusted) {
			return true
		}
	}
	return false
}

func GetRealClientIP() gin.HandlerFunc {
	return func(c *gin.Context) {
		var clientIP string

		if len(trustedProxyIPs) > 0 {
			xff := c.GetHeader("X-Forwarded-For")
			xri := c.GetHeader("X-Real-IP")

			if xff != "" {
				parts := strings.Split(xff, ",")
				for i := len(parts) - 1; i >= 0; i-- {
					ipStr := strings.TrimSpace(parts[i])
					if ipStr == "" {
						continue
					}
					ip := net.ParseIP(ipStr)
					if ip == nil {
						continue
					}
					if isTrustedProxy(ip) {
						continue
					}
					clientIP = ipStr
					break
				}
			}

			if clientIP == "" && xri != "" {
				ip := net.ParseIP(strings.TrimSpace(xri))
				if ip != nil && !isTrustedProxy(ip) {
					clientIP = strings.TrimSpace(xri)
				}
			}
		}

		if clientIP == "" {
			clientIP = c.ClientIP()
		}

		if clientIP == "" {
			clientIP = "unknown"
		}

		c.Set("real_client_ip", clientIP)
		c.Request.Header.Set("X-Real-IP", clientIP)
		c.Next()
	}
}

func GetClientIP(c *gin.Context) string {
	if ip, exists := c.Get("real_client_ip"); exists {
		if clientIP, ok := ip.(string); ok && clientIP != "" {
			return clientIP
		}
	}
	return c.ClientIP()
}
