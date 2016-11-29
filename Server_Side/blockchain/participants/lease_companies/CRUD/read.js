/*eslint-env node */

var tracing = require(__dirname+'/../../../../tools/traces/trace.js');
var reload = require('require-reload')(require),
    participants = reload(__dirname+'/../../participants_info.js');

var read = function(req, res)
{
	participants = reload(__dirname+'/../../participants_info.js');
	tracing.create('ENTER', 'GET blockchain/participants/airports', {});
	
	if(!participants.participants_info.hasOwnProperty('airports'))
	{
		res.status(404)
		var error = {}
		error.message = 'Unable to retrieve lease companies';
		error.error = true;
		tracing.create('ERROR', 'GET blockchain/participants/airports', error);
		res.send(error)
	} 
	else
	{
		tracing.create('EXIT', 'GET blockchain/participants/airports', {"result":participants.participants_info.airports});
		res.send({"result":participants.participants_info.airports})
	}
}
exports.read = read;