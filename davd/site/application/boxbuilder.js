dependsOn(
	"../framework/jquery.js",
	"../framework/gridster/gridster.js");

(function(global){
	var boxbuilder = global.boxbuilder || (global.boxbuilder = {});
	boxbuilder.createBoxAt = function (grid, html) {
		if (!html) { html = boxbuilder.newEmptyBox(); }
		grid.add_widget(html, 1, 1);
	};
	boxbuilder.newEmptyBox = function() {
		return $('<div class="box invert-fill"><div class="header"><span class="label">Empty box</span></div><div class="content"></div></div>');
	};
	$(function(){});
}(this));
