// Development-only visual harness. Never loaded by the production entry point.
import React, {useState,useCallback} from 'react';
import {createRoot} from 'react-dom/client';
import {AppShell} from './App';
import {AppContext, AppContextValue} from './context/AppContext';
import './style.css';
function Preview(){
 const [active,setActive]=useState<any>('dashboard');
 const [toast,setToast]=useState<any>(null);
 const unavailable=useCallback(async()=>setToast({text:"Изолированная проверка интерфейса: сетевые действия отключены",variant:"info"}),[]);
 const emptyLogs={coreLogTail:'',dpiLogTail:'',coreLogPath:'',dpiLogPath:'',coreErrors:'',dpiErrors:'',coreConnLines:'',dpiConnLines:''};
 const value=new Proxy({active,setActive,toast,setToast,busy:false,globalBusy:null,proxyEnabled:false,dpiEnabled:false,proxyPhase:'idle',appVersion:'0.7.4',health:{proxyLocalOk:false,dpiRunning:false},uiState:{proxyPort:2080,proxyMode:'system_proxy',routingProfile:'selective-proxy',dpiProfiles:[],licenseTier:'free'},rt:{downloadBps:0,uploadBps:0,sessionElapsedSeconds:0},speedHistory:[],sortedServers:[],routingProfilesUnique:[{name:'selective-proxy'}],activeRoutingProfileMeta:{final:'direct'},logs:emptyLogs,dpiTuneResult:null,updateInfo:null,updateHint:'',guideText:'',diagText:''},{get:(target:any,key:string)=>key in target ? target[key] : unavailable}) as AppContextValue;
 return <AppContext.Provider value={value}><AppShell/></AppContext.Provider>;
}
if(import.meta.env.DEV) createRoot(document.getElementById('root')!).render(<Preview/>);
