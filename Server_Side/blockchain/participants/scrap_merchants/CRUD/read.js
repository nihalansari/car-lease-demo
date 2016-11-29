/*eslint-env node */

var tracing = require(__dirname+'/../../../../tools/traces/trace.js');
var reload = require('require-reload')(require),
	participants = reload(__dirname+'/../../participants_info.js');

var read = function(req, res)
{
	participants = reload(__dirname+'/../../participants_info.js');
	tracing.create('ENTER', 'GET blockchain/participants/delivery', {});
	
	if(!participants.participants_info.hasOwnProperty('delivery'))
	{
		res.status(404)
		var error = {}
		error.message = 'Unable to retrieve delivery firms';
		error.error = true;
		tracing.create('ERROR', 'GET blockchain/participants/delivery', error);
		res.send(error)
	} 
	else
	{
		tracing.create('EXIT', 'GET blockchain/participants/delivery', {"result":participants.participants_info.delivery});
		res.send({"result":participants.participants_info.delivery})
	}
}
exports.read = read;