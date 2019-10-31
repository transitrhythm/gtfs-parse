package main

// Prelude -
var Prelude = `
<!DOCTYPE html>
<html lang="en-US">
<head>
	<meta charset="UTF-8">
	<title>Schedule Adherence for a series of Transit Trips Displayed Dynamically</title>
	<style>
		.grid {
			-webkit-column-count: %d; /* Old Chrome, Safari and Opera */
			-moz-column-count: %d; /* Old Firefox */
			column-count: %d;
		}
	</style>
	<script type ="text/javascript" src ="adherence.js"></script>
</head>	
<body onload="init()">
	<h1><u>Schedule Adherence</u></h1><hr>`

// Epilog -
var Epilog = `
<div class="grid">
<script>
setInterval(function() {
	var imgs = document.getElementsByTagName("IMG");
	for (var i=0; i < imgs.length; i++) {
		var eqPos = imgs[i].src.lastIndexOf("=");
		var src = imgs[i].src.substr(0, eqPos+1);
		imgs[i].src = src + Math.random();
	}
}, %d);
</script>
</div>
</body>
</html>`

// HTMLDropdown
var HTMLDropdown = `
	<style>
	.grid1 {
		-webkit-column-count: 2; /* Old Chrome, Safari and Opera */
		-moz-column-count: 2; /* Old Firefox */
		column-count: 2;
	}
	table, th, td {
		border: 1px solid black;
		border-collapse: collapse;
	  }
	th,td { padding: 20px }
	agency { display: none }
	div.hidden { visibility: hidden; }
	</style>
	<!--<div class="grid1">-->
	<table>
	<tr>
	<th>
	<div id="authority">
	<h2>Transit Authority</h2>
	<p><i>Change the authority using the drop-down list:</i></p>	
	<form action="/authority_page.php">
	  <select name="authority">
	  <option value="1">BC Transit</option>
	  <option value="2">Vancouver Translink</option>
	  </select>
	  <br><br>
	  <input type="submit">
	</form>	
	</div>
	</th>
	<th>
	<div id="agency" class="hidden">
	<h2>Transit Agency</h2>
	<p><i>Change the transit agency using the drop-down list:</i></p>	
	<form action="/agency_page.php">
	  <select name="agency">
		<option value="12">Comox Valley Transit System</option>
		<option value="8">Kamloops Transit System</option>
		<option value="7">Kelowna Regional Transit System</option>
		<option value="5">RDN Transit System</option>
		<option value="4">Squamish Transit System</option>
		<option value="1">Victoria Regional Transit System</option>
		<option value="3">Whistler Transit System</option>
	  </select>
	  <br><br>
	  <input type="submit">
	</form>
	</div>
	<!--</div>-->
	</th>
	</tr>
	</table>`
