package handler

import "github.com/gofiber/fiber/v2"

// RedocHandler serves the ReDoc API documentation UI
func RedocHandler() fiber.Handler {
	html := `<!DOCTYPE html>
<html>
<head>
  <title>XBank API Documentation</title>
  <meta charset="utf-8"/>
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <style>body { margin: 0; padding: 0; }</style>
</head>
<body>
  <redoc spec-url="/swagger/doc.json"
         hide-download-button
         theme='{
           "colors": { "primary": { "main": "#1a56db" } },
           "typography": { "fontSize": "15px", "fontFamily": "Inter, sans-serif" }
         }'>
  </redoc>
  <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
</body>
</html>`

	return func(c *fiber.Ctx) error {
		c.Set("Content-Type", "text/html")
		return c.SendString(html)
	}
}
