dependsOn(
	"../framework/jquery.js",
	"../framework/gridster/gridster.js",
	"./context.js");
(function(global){
	$(function(){
		global.context.init();
	});
}(this));
