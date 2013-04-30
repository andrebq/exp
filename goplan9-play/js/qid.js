
	exports.Qid = function() {
		this.Path = exports.Uint64.fromInteger(0);
		this.Vers = 0;
		this.Type = 0;
	};

	exports.Qid.prototype.asString = function() {
		var t = "";
		if (this.Type & exports.QTDIR) {
			t += "d";
		}
		if (this.Type & exports.QTAPPEND) {
			t += "a";
		}
		if (this.Type & exports.QTEXCL) {
			t += "l";
		}
		if (this.Type & exports.QTAUTH) {
			t += "A";
		}
		return sprintf("(%s %d %s)", this.Path.asString(), this.Vers, t);
	};
