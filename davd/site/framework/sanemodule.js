(function(global){
	function importFn() {
		var args = Array.prototype.slice.apply(arguments);
		var sz = args.length;
		for(var pos = 0; pos < sz; pos++) {
			console.log("dependsOn: " + args[pos]);
		}
	}
	global.dependsOn = global.dependsOn || importFn;
}(this));
