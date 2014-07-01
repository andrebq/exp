function PandoraCtrl($scope) {
	// makeNewField create a empty field to be used by the user
	// to include a new field in the message
	function makeNewField() {
		return {
			caption: "",
			name: "",
			initialValue: ""};
	}

	function fetchMessages() {
		var toEncode = {
			receiver: "b@remote",
			receivedat: "-24h",
		};
		$.getJSON("/api/admin/headers", $.param(toEncode), function(result){
			$scope.$apply(function(){
				_.each(result, function(value){
					$scope.msgsSent.push(mergeKeys(value));
				});
			});
		});
	}

	function mergeKeys(msg) {
		var out = {};
		_.each(msg, function(value, key){
			out[key] = _.reduce(value, function(acc, val) {
				return acc += val;
			}, "");
		});
		return out;
	}

	// validField returns true when fld have the required fields filled;
	function validField(fld) {
		return !!fld.name && !!fld.caption;
	}

	// encodeForPandora return the msg (which is a list of fields), 
	// in the format expected by the Pandora Web Service
	function encodeForPandora(msg) {
		var toEncode = {};
		$.each(msg.fields, function(i, fld){
			toEncode[fld.name] = fld.value;
		});
		return $.param(toEncode);
	}

	// clearMessage will reset all fields on msg to the respective
	// initial value
	function clearMessage(msg) {
		$.each(msg.fields, function(i, fld) {
			fld.value = fld.initialValue;
		});
	}

	$scope.newField = makeNewField();
	$scope.msgsSent = [];

	$scope.message = {
		fields: [
			{ caption:"To", name: "receiver", initialValue: "" },
			{ caption:"From", name: "sender", initialValue: "" },
			{ caption:"Delay", name: "delay", initialValue: "", value: "1s" },
			{ caption:"Topic", name: "topic", initialValue: "", value: "no subject" }, 
		]
	};

	$scope.addField = function() {
		if (validField($scope.newField)) {
			$scope.message.fields.push({
				caption: $scope.newField.caption,
				name: $scope.newField.name,
				value: $scope.newField.initialValue,
				initialValue: $scope.newField.initialValue,
			});
			$scope.newField = makeNewField();
		}
	};

	$scope.sendMessage = function() {
		var encoded = encodeForPandora($scope.message);
		$.post("/api/send", encoded, function(response){
			console.log(response);
		}, 'text');
		return true;
	};

	fetchMessages();
}
