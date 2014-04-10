(function(global){

function resize(element) {
	// run the resize more than one time
	var pendingResizes = 2;
	var realResize = function() {
		var w = element.offsetWidth,
			h = element.offsetHeight;
		if (w == 0 || h == 0) {
			// wait for the next cycle
			window.requestAnimationFrame(realResize);
			return;
		}
		var cm = element.children[0].CodeMirror;
		cm.setSize(w, h);
		cm.refresh();
		pendingResizes--;
		if (pendingResizes > 0) {
			window.requestAnimationFrame(realResize);
		}
	}
	window.requestAnimationFrame(realResize)
}

function autosize() {
	var elements = document.querySelectorAll(".gocm-editor");
	var sz = elements.length;
	for(var idx = 0; idx < sz; idx++) {
		el = elements[idx];
		resize(el);
	}
}

if (global.addEventListener) {
	global.addEventListener("resize", autosize, false);
}

global.GoCM = {};
global.GoCM.autosize = autosize;
global.GoCM.createEditorAt = function(el, autosize) {
	while(el.children.length !== 0) {
		el.removeChild(el.children[0]);
	}
	var cm = CodeMirror(el);
	el.classList.add("gocm-editor");
	if (autosize) {
		autosize();
	}
};

global.GoCM.editorOf = function(el) {
	if (el.classList.contains("gocm-editor")) {
		return el.children[0].CodeMirror;
	} else if (el.CodeMirror !== undefined) {
		return el.CodeMirror;
	} else {
		return null;
	}
};

}(window));
