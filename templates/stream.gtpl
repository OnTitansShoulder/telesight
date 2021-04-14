<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>Webcam Streams</title>
	</head>
	<body>
		<h1>Welcome to the streaming page!</h1>
        <p>Select a stream source:</p>
		<select id="stream-selection" onchange="loadVideo()">
			{{ range .StreamSources }}<option value={{ .StreamURL }}>{{ .DisplayText }}</option>{{ end }}
		</select>
        <div id="video-container" style="text-align: center">
            
        </div>
	</body>
</html>
<script>
function loadVideo() {
	selected = document.getElementById("stream-selection");
	videoContainer = document.getElementById("video-container");
	videoContainer.innerHTML = "<img width='640px' height='360px' style='width: 640px; height: 360px' src='" + selected.value + "'></video>";
}
</script>