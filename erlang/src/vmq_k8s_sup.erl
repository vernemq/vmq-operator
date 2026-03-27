%%%-------------------------------------------------------------------
%% @doc vmq_k8s top level supervisor.
%% @end
%%%-------------------------------------------------------------------

-module(vmq_k8s_sup).

-behaviour(supervisor).

%% API
-export([start_link/0]).

%% Supervisor callbacks
-export([init/1]).

-define(SERVER, ?MODULE).

%%====================================================================
%% API functions
%%====================================================================

start_link() ->
    supervisor:start_link({local, ?SERVER}, ?MODULE, []).

%%====================================================================
%% Supervisor callbacks
%%====================================================================

%% Child :: {Id,StartFunc,Restart,Shutdown,Type,Modules}
init([]) ->
    Reloader = {vmq_k8s_reloader, {vmq_k8s_reloader, start_link, []},
                permanent, 5000, worker, [vmq_k8s_reloader]},

    {ok, { {one_for_all, 0, 1}, [Reloader]} }.

%%====================================================================
%% Internal functions
%%====================================================================
