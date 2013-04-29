test("asString", function(){

	var call = new exports.Fcall();
	call.Type = exports.MessageType.Tversion;
	call.Version = "9p2000.u";
	call.Tag = exports.NOTAG;
	call.Msize = 0;

	ok("Tversion tag 65535 msize 0 version '9p2000.u'" === call.asString(), "Tversion asString");
	call.Type = exports.MessageType.Rversion;
	ok("Rversion tag 65535 msize 0 version '9p2000.u'" === call.asString(), "Rversion asString");

});
