(function(global){

function resize(element) {
	var w = element.offsetWidth,
		h = element.offsetHeight;
	if (w == 0 || h == 0) {
		return;
		// wait for the next cycle
		window.requestAnimationFrame(autosize);
	}
	var cm = element.children[0].CodeMirror;
	cm.setSize(w, h);
	cm.refresh();
}

function autosize() {
	var elements = document.querySelectorAll(".gocm-editor");
	var sz = elements.length;
	for(var idx = 0; idx < sz; idz++) {
		el = elements[idx];
		resize(el);
	}
}

if (global.addEventListener) {
	global.addEventListener("resize", autosize, false);
}

global.GoCM = {};
global.GoCM.autosize = autosize;

}(window));
