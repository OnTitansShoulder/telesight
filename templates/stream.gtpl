<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<meta name="viewport" content="width=device-width, initial-scale=1">
    	<meta name="description" content="">
		<title>Webcam Streams</title>
		<style>
			.bd-placeholder-img {
				font-size: 1.125rem;
				text-anchor: middle;
				-webkit-user-select: none;
				-moz-user-select: none;
				user-select: none;
			}

			@media (min-width: 768px) {
				.bd-placeholder-img-lg {
					font-size: 3.5rem;
				}
			}

			.icon-list {
				padding-left: 0;
				list-style: none;
			}
			.icon-list li {
				display: flex;
				align-items: flex-start;
				margin-bottom: .25rem;
			}
			.icon-list li::before {
				display: block;
				flex-shrink: 0;
				width: 1.5em;
				height: 1.5em;
				margin-right: .5rem;
				content: "";
				background: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' fill='%23212529' viewBox='0 0 16 16'%3E%3Cpath d='M8 0a8 8 0 1 1 0 16A8 8 0 0 1 8 0zM4.5 7.5a.5.5 0 0 0 0 1h5.793l-2.147 2.146a.5.5 0 0 0 .708.708l3-3a.5.5 0 0 0 0-.708l-3-3a.5.5 0 1 0-.708.708L10.293 7.5H4.5z'/%3E%3C/svg%3E") no-repeat center center / 100% auto;
			}
		</style>
        <link href="https://cdn.jsdelivr.net/npm/bootstrap@5.0.0-beta3/dist/css/bootstrap.min.css" rel="stylesheet" integrity="sha384-eOJMYsd53ii+scO/bJGFsiCZc+5NDVN2yr8+0RDqr0Ql0h+rP48ckxlpbzKgwra6" crossorigin="anonymous">
	</head>
	<body>
		<div class="col-lg-8 mx-auto p-3 py-md-5">
			<header class="d-flex align-items-center pb-3 mb-5 border-bottom">
				<a href="/" class="d-flex align-items-center text-dark text-decoration-none">
					<span class="fs-4">Streaming From Cams</span>
				</a>
				<nav class="d-inline-flex mt-2 mt-md-0 ms-md-auto">
                    <a class="me-3 py-2 text-dark text-decoration-none btn btn-light" href="/telesight/stream/">Stream</a>
                    <a class="me-3 py-2 text-dark text-decoration-none btn btn-light" href="/telesight/watch/">Watch</a>
                    <a class="me-3 py-2 text-dark text-decoration-none btn btn-light" href="/telesight/videos/">Local Videos</a>
                </nav>
			</header>
			<div style="display: flex;">
				<h3>Select a stream source:</h3>
				<select id="stream-selection" onchange="loadVideo()" style="margin-left: 1em;">
					<option value="">None-selected</option>
					{{ range .StreamSources }}<option value={{ .URL }}>{{ .DisplayText }}</option>{{ end }}
				</select>
			</div>

			<hr class="col-3 col-md-2 mb-5">

			<div id="video-container" style="text-align: center"></div>
		</div>
	</body>
</html>
<script>
function loadVideo() {
	var selected = document.getElementById("stream-selection");
	videoContainer = document.getElementById("video-container");
    if ( selected.value.length == 0 ) {
        videoContainer.innerHTML = ""
        return
    }
	videoContainer.innerHTML = "<img style='width: 100%; border: solid  ' src='" + selected.value + "'/>";
}
</script>
<script src="https://cdn.jsdelivr.net/npm/bootstrap@5.0.0-beta3/dist/js/bootstrap.bundle.min.js" integrity="sha384-JEW9xMcG8R+pH31jmWH6WWP0WintQrMb4s7ZOdauHnUtxwoG2vI5DkLtS3qm9Ekf" crossorigin="anonymous"></script>
