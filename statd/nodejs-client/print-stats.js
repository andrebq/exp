var readline = require('readline');
function ProcessEvent(ev) {
	process.stderr.write("ev: " + JSON.stringify(ev) + "\n");
}

process.stdin.setEncoding('utf-8');
var rl = readline.createInterface({
	input: process.stdin,
	output: process.stderr,
	terminal: false,
});

rl.on('line', function(line){
	if (line.trim() === "" ) { return; }
	try {
		ProcessEvent(JSON.parse(line.trim()));
	} catch(ex) {
		process.stderr.write("line: " + line + "\n");
		process.stderr.write("error: " + ex + "\n");
		process.exit(1)
	}
})

process.stdin.on('end', function(data){
	process.exit(0);
});

var lastId = 0;

if (process.argv.length > 2) {
	try {
		lastId = parseInt(process.argv[2]);
	} catch(err) {
		process.stderr.write("error: " + ex);
		lastId = 0;
	}
}

process.stdout.write(lastId + "\n");
process.stdin.resume();
