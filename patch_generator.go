package main

import (
	"os"
	"strings"
)

func main() {
	b, _ := os.ReadFile("internal/docker/generator.go")
	s := string(b)
	
	oldStr := `COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next/standalone ./
COPY --from=builder --chown=nextjs:nodejs /app/.next/static ./.next/static

USER nextjs

EXPOSE 3000
ENV PORT 3000

CMD ["node", "server.js"]`

	newStr := `COPY --from=builder /app/public ./public
COPY --from=builder --chown=nextjs:nodejs /app/.next ./.next
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/package.json ./package.json

USER nextjs

EXPOSE 3000
ENV PORT 3000

CMD ["npm", "start"]`

	s = strings.Replace(s, oldStr, newStr, 1)
	os.WriteFile("internal/docker/generator.go", []byte(s), 0644)
}
