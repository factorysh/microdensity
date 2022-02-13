package application

import "io"

func writeHTMLHeader(w io.Writer) {
	w.Write([]byte(`
	<!doctype html>
	<html>
	  <head>
	    <meta charset="UTF-8">
	    <meta name="viewport" content="width=device-width, initial-scale=1.0">
		<script src="https://cdn.tailwindcss.com"></script>
		<title>µdensity</title>
	  </head>
	  <body>
	`))
}

func writeHTMLFooter(w io.Writer) {
	w.Write([]byte(`
	  </body>
	</html>
	`))
}
