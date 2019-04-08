%%%-------------------------------------------------------------------
%%% @author graf
%%% @copyright (C) 2019, graf
%%% @doc
%%%
%%% @end
%%% Created : 2019-04-02 20:09:28.354956
%%%-------------------------------------------------------------------
-module(vmq_k8s_reloader).

-behaviour(gen_server).

%% API
-export([start_link/0]).

%% gen_server callbacks
-export([init/1,
         handle_call/3,
         handle_cast/2,
         handle_info/2,
         terminate/2,
         code_change/3]).

-define(SERVER, ?MODULE).

-record(state, {map=#{}, clustering_state=[]}).

%%%===================================================================
%%% API
%%%===================================================================

%%--------------------------------------------------------------------
%% @doc
%% Starts the server
%%
%% @spec start_link() -> {ok, Pid} | ignore | {error, Error}
%% @end
%%--------------------------------------------------------------------
start_link() ->
    gen_server:start_link({local, ?SERVER}, ?MODULE, [], []).

%%%===================================================================
%%% gen_server callbacks
%%%===================================================================

%%--------------------------------------------------------------------
%% @private
%% @doc
%% Initializes the server
%%
%% @spec init(Args) -> {ok, State} |
%%                     {ok, State, Timeout} |
%%                     ignore |
%%                     {stop, Reason}
%% @end
%%--------------------------------------------------------------------
init([]) ->
    erlang:send_after(1000, self(), check_config),
    {ok, #state{}}.

%%--------------------------------------------------------------------
%% @private
%% @doc
%% Handling call messages
%%
%% @spec handle_call(Request, From, State) ->
%%                                   {reply, Reply, State} |
%%                                   {reply, Reply, State, Timeout} |
%%                                   {noreply, State} |
%%                                   {noreply, State, Timeout} |
%%                                   {stop, Reason, Reply, State} |
%%                                   {stop, Reason, State}
%% @end
%%--------------------------------------------------------------------
handle_call(_Request, _From, State) ->
    Reply = ok,
    {reply, Reply, State}.

%%--------------------------------------------------------------------
%% @private
%% @doc
%% Handling cast messages
%%
%% @spec handle_cast(Msg, State) -> {noreply, State} |
%%                                  {noreply, State, Timeout} |
%%                                  {stop, Reason, State}
%% @end
%%--------------------------------------------------------------------
handle_cast(_Msg, State) ->
    {noreply, State}.

%%--------------------------------------------------------------------
%% @private
%% @doc
%% Handling all non call/cast messages
%%
%% @spec handle_info(Info, State) -> {noreply, State} |
%%                                   {noreply, State, Timeout} |
%%                                   {stop, Reason, State}
%% @end
%%--------------------------------------------------------------------

handle_info(check_config, #state{map=ConfigState0, clustering_state=ClusteringState0} = State0) ->
    ConfigFile = os:getenv("VMQ_CONFIGMAP", "/vernemq/etc/vernemq.yaml"),
    ConfigState1 =
    try yamerl_constr:file(ConfigFile) of
        Config ->
            apply_config(Config, ConfigState0)
    catch
        throw:{yamerl_exception, Exception} ->
            lager:error("Can't parse YAML ConfigMap ~p", [Exception]),
            ConfigState0;
        E:R ->
            lager:error("Error while parsing ConfigMap ~p ~p", [E, R]),
            ConfigState0
    end,
    ClusterviewFile = os:getenv("VMQ_CLUSTERVIEW", "/vernemq/etc/configmaps/clusterview/clusterview.yaml"),
    Ret =
    case file:read_file(ClusterviewFile) of
        {ok, Content} ->
            Nodes = [N || N <- re:split(Content, ";"), N =/= <<>>],
            check_clustering(Nodes, ClusteringState0);
        {error, Reason} ->
            lager:error("Can't read Clusterview File ~p due to ~p", [ClusterviewFile, Reason]),
            ClusteringState0
    end,
    case Ret of
        its_over ->
            {stop, normal, State0};
        _ ->
            erlang:send_after(1000, self(), check_config),
            {noreply, State0#state{map=ConfigState1, clustering_state=Ret}}
    end.

apply_config([Config|_], CurrentState) ->
    State0 = #{},
    State1 = apply_plugins_config(proplists:get_value("plugins", Config, []), State0, CurrentState),
    State2 = apply_listener_config(proplists:get_value("listeners", Config, []), State1, CurrentState),
    State3 = apply_value_config(proplists:get_value("config", Config, []), State2, CurrentState),
    State3.

apply_plugins_config([PluginConfig|Rest], Acc, CurrentState) ->
    case proplists:get_value("name", PluginConfig) of
        undefined ->
            lager:error("Can't apply plugin config, missing 'name' in ~p", [PluginConfig]),
            apply_plugins_config(Rest, Acc, CurrentState);
        Name ->
            case maps:is_key({plugin, Name}, CurrentState) of
                true ->
                    % already installed
                    apply_plugins_config(Rest, maps:put({plugin, Name}, enabled, Acc), CurrentState);
                false ->
                    Acc1 =
                    case proplists:get_value("path", PluginConfig) of
                        undefined ->
                            command(["plugin", "enable", "-n", Name],
                                    succf({plugin, Name}, enabled), Acc);
                        Path ->
                            command(["plugin", "enable", "-n", Name, "-p", Path],
                                    succf({plugin, Name}, enabled), Acc)
                    end,
                    apply_plugins_config(Rest, Acc1, CurrentState)
            end
    end;
apply_plugins_config([], NewState, OldState) ->
    New = [Name || {plugin, Name} <- maps:keys(NewState)],
    Old = [Name || {plugin, Name} <- maps:keys(OldState)],
    ToBeDisabled = Old -- New,
    lists:foreach(fun(Name) ->
                          command(["plugin", "disable", "-n", Name])
                  end, ToBeDisabled),
    NewState;
apply_plugins_config(null, NewState, _OldState) ->
    %% empty list is decoded as null
    NewState.

apply_listener_config([ListenerConfig|Rest], Acc, State) ->
    case {proplists:get_value("address", ListenerConfig),
          proplists:get_value("port", ListenerConfig)} of
        {Addr, IPort} when Addr =/= undefined, IPort =/= undefined ->
            Port = integer_to_list(IPort),
            Flags0 = lists:foldl(fun({K, V}, CAcc) when K =/= "address", K =/= "port", K =/= "type" ->
                                         case V of
                                             "on" ->
                                                 % flag
                                                 ["--" ++ K | CAcc];
                                             "off" ->
                                                 % flag
                                                 CAcc;
                                             _ ->
                                                 ["--" ++ K ++ "=" ++ V | CAcc]
                                         end;
                                    (_, CAcc) ->
                                         CAcc
                                 end, [], ListenerConfig),
            Flags1 =
            case proplists:get_value("type", ListenerConfig) of
                "mqtt" -> Flags0;
                "mqtts" -> ["--ssl" | Flags0];
                "ws" -> ["--websocket" | Flags0];
                "wss" -> ["--ssl", "--websocket" | Flags0];
                "http" -> ["--http" | Flags0];
                "https" -> ["--ssl", "--http" | Flags0];
                _ ->
                    Flags0
            end,
            Flags = lists:usort(Flags1),
            case maps:get({listener, {Addr, Port}}, State, undefined) of
                Flags ->
                    % no change
                    apply_listener_config(Rest, maps:put({listener, {Addr, Port}}, Flags, Acc), State);
                _UndefOldFlags ->
                    % delete
                    DeleteCommand = ["listener", "delete", "address=" ++ Addr, "port=" ++ Port],
                    command(DeleteCommand),
                    StartCommand = ["listener", "start", "address=" ++ Addr, "port=" ++ Port] ++ Flags,
                    Acc1 = command(StartCommand, succf({listener, {Addr, Port}}, Flags), Acc),
                    apply_listener_config(Rest, Acc1, State)
            end;
        _ ->
            lager:error("address or port not set in ~p", [ListenerConfig]),
            apply_listener_config(Rest, Acc, State)
    end;
apply_listener_config([], NewState, OldState) ->
    New = [AddrPort || {listener, AddrPort} <- maps:keys(NewState)],
    Old = [AddrPort || {listener, AddrPort} <- maps:keys(OldState)],
    ToBeDeleted = Old -- New,
    lists:foreach(fun({Addr, Port}) ->
                          command(["listener", "delete", "address=" ++ Addr, "port=" ++ Port])
                  end, ToBeDeleted),
    NewState;
apply_listener_config(null, NewState, _OldState) ->
    %% empty list is decoded as null
    NewState.

apply_value_config([{ConfigKey, ConfigValue}|Rest], Acc, CurrentState) ->
    case maps:get({config, ConfigKey}, CurrentState, undefined) of
        undefined ->
            case default_val(ConfigKey) of
                {ok, Default} ->
                    Acc1 = command(["set", ConfigKey ++ "=" ++ ConfigValue],
                                   succf({config, ConfigKey}, {ConfigValue, Default}), Acc),
                    apply_value_config(Rest, Acc1, CurrentState);
                {error, invalid_key} ->
                    lager:error("Invalid config key ~p", [ConfigKey]),
                    apply_value_config(Rest, Acc, CurrentState)
            end;
        {ConfigValue, Default} ->
            % same value
            apply_value_config(Rest, maps:put({config, ConfigKey}, {ConfigValue, Default}, Acc), CurrentState);
        {_Other, Default} ->
            Acc1 = command(["set", ConfigKey ++ "=" ++ ConfigValue],
                           succf({config, ConfigKey}, {ConfigValue, Default}), Acc),
            apply_value_config(Rest, Acc1, CurrentState)
    end;
apply_value_config([], NewState, OldState) ->
    % TODO: similar to plugins, we have to reset to defaults...
    % currently no 'easy' way to figure out the default value
    New = [ConfigKey || {config, ConfigKey} <- maps:keys(NewState)],
    Old = [ConfigKey || {config, ConfigKey} <- maps:keys(OldState)],
    ToBeReset = Old -- New,
    lists:foreach(fun(ConfigKey) ->
                          {_, DefaultValue} = maps:get(ConfigKey, OldState),
                          command(["set", ConfigKey ++ "=" ++ DefaultValue])
                  end, ToBeReset),
    NewState.

default_val(ConfigKey) ->
    case clique_config:show([ConfigKey], []) of
        [{table, [Res]}] ->
            {ok, proplists:get_value(ConfigKey, Res)};
        {error, {invalid_config_keys, _}} ->
            {error, invalid_key}
    end.

check_clustering(CurNodes, OldNodes) ->
    MySelf = atom_to_binary(node(), utf8),
    case {lists:member(MySelf, CurNodes),
          lists:member(MySelf, OldNodes)} of
        {true, false} ->
            case CurNodes -- [MySelf] of
                [] ->
                    % we're the FirstNode, others will join
                    CurNodes;
                [FirstNode|_] ->
                    % join with FirstNode
                    command(["cluster", "join", "discovery-node=" ++ binary_to_list(FirstNode)]),
                    CurNodes
            end;
        {false, true} ->
            % we have to leave, this will teardown the mqtt listeners and init:stop the node when finished
            command(["cluster", "leave" "node=" ++ binary_to_list(MySelf), "--timeout=3600", "--kill_sessions"]),
            its_over;
        _ ->
            % no change
            CurNodes
    end.

succf(K,V) ->
    fun(A) -> maps:put(K, V, A) end.

command(Args) ->
    command(Args, fun(_) -> ignore end, ignore).

command(Args, SuccessFun, Acc) ->
    Cmd = ["vmq-admin" | Args],
    try vmq_server_cli:command(Cmd, false) of
        {ok, Ret} ->
            lager:info("Execute: ~p ~p", [Cmd, proplists:get_value(text, Ret, "Done")]),
            SuccessFun(Acc);
        {error, [{alert, [{text, Txt}]}]} ->
            % special case for Plugin
            Text = lists:flatten(Txt),
            case string:find(Text, "already_enabled") of
                nomatch ->
                    lager:error("Execute error: ~p ~p", [Cmd, Text]),
                    Acc;
                _ ->
                    lager:info("Execute: ~p ~p", [Cmd, Text]),
                    SuccessFun(Acc)
            end;
        {error, Error} ->
            lager:error("Execute error: ~p ~p", [Cmd, Error]),
            Acc;
        Other ->
            lager:error("Execute error: ~p ~p", [Cmd, Other]),
            Acc
    catch
        E:R ->
            lager:error("Execute error: ~p ~p ~p", [Cmd, E, R]),
            Acc
    end.

%%--------------------------------------------------------------------
%% @private
%% @doc
%% This function is called by a gen_server when it is about to
%% terminate. It should be the opposite of Module:init/1 and do any
%% necessary cleaning up. When it returns, the gen_server terminates
%% with Reason. The return value is ignored.
%%
%% @spec terminate(Reason, State) -> void()
%% @end
%%--------------------------------------------------------------------
terminate(_Reason, _State) ->
    ok.

%%--------------------------------------------------------------------
%% @private
%% @doc
%% Convert process state when code is changed
%%
%% @spec code_change(OldVsn, State, Extra) -> {ok, NewState}
%% @end
%%--------------------------------------------------------------------
code_change(_OldVsn, State, _Extra) ->
        {ok, State}.

%%%===================================================================
%%% Internal functions
%%%===================================================================




