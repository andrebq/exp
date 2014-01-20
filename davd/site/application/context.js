dependsOn(
	"../framework/jquery.js",
	"../framework/gridster.gridster.js");

(function(global){
	// only one context needed
	if (global.context) { return; }
	global.context = {}
	var grid;
	function createGrid() {
		grid = $(".gridster").gridster({
			widget_margins: [10, 10],
			widget_base_dimensions: [200, 200],
			resize: {
				enabled: true
			}
		}).data("gridster");
		console.log("grid created");
	}
	function getGrid() {
		return grid;
	}
	function getDoc() {
		return global.document;
	}
	global.context.getGrid = getGrid;
	global.context.getDoc = getDoc;
	global.context.init = function() {
		createGrid();
	};
}(this));
