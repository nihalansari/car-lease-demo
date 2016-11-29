/*eslint-env node*/

var tracing = require(__dirname+'/../../../../tools/traces/trace.js');
var reload = require('require-reload')(require),
	participants = reload(__dirname+'/../../participants_info.js');

var read = function(req, res)
{
	participants = reload(__dirname+'/../../participants_info.js');
	
	tracing.create('ENTER', 'GET blockchain/participants/suppliers', {});

	if(!participants.participants_info.hasOwnProperty('suppliers'))
	{
		res.status(404)
		var error = {}
		error.message = 'Unable to retrieve suppliers';
		error.error = true;
		tracing.create('ERROR', 'GET blockchain/participants/suppliers', error);
		res.send(error)
	} 
	else
	{
		tracing.create('EXIT', 'GET blockchain/participants/suppliers', {"result":participants.participants_info.suppliers});
		res.send({"result":participants.participants_info.suppliers})
	}

}
exports.read = read;