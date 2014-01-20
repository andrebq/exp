dependsOn(
	"../framework/jquery.js",
	"../framework/jquery.hotkeys.js",
	"./boxbuilder.js",
	"./context.js");
(function(global){
	var shortcuts = global.shortcuts || (global.shortcuts = {});
	shortcuts.add = function(ctx, comb, handler) {
		$(ctx.getDoc()).bind('keydown', comb, function(){
			handler();
			return false;
		});
	};
	$(function(){
		shortcuts.add(global.context, 'f4', create_empty_box);
	});

	function create_empty_box() {
		boxbuilder.createBoxAt(global.context.getGrid());
	}
}(this));
