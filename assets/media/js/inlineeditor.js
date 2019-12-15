var inlineeditor_isIE=(navigator.userAgent.toLowerCase().search('msie')!=-1&&navigator.userAgent.toLowerCase().search('opera')==-1)?true:false;var InlineEditor={alreadyInited:false,init:function(arg)
{var isNode=false;var isEvent=false;if(arg.nodeType)
isNode=true;if(!isNode&&InlineEditor.alreadyInited)
return;var rootEl=isNode?arg:document;var allEl=rootEl.getElementsByTagName('*');for(var i=0;i<allEl.length;i++){if(InlineEditor.checkClass(allEl[i],'editable')){switch(allEl[i].tagName){case'TABLE':var tds=allEl[i].getElementsByTagName('td');for(var j=0;j<tds.length;j++)
InlineEditor.recursiveAddOnClickHandler(tds[j]);break;default:InlineEditor.recursiveAddOnClickHandler(allEl[i]);}}}
if(!isNode)
InlineEditor.alreadyInited=true;},addClass:function(o,c){return InlineEditor.jscss('add',o,c);},removeClass:function(o,c){return InlineEditor.jscss('remove',o,c);},checkClass:function(o,c){return InlineEditor.jscss('check',o,c);},swapClass:function(o,c1,c2){return InlineEditor.jscss('swap',o,c1,c2);},columnNumber:function(cell)
{if(cell.nodeType!=1)return-1;if(cell.tagName!='TD')return-1;if(!cell.parentNode||cell.parentNode.tagName!='TR')return-1;var tr=cell.parentNode;var tds=tr.getElementsByTagName('TD');for(var i=0;i<tds.length;i++)
if(tds[i]==cell)
return i;return-1;},rowNumber:function(cell)
{if(cell.nodeType!=1)return-1;if(cell.tagName!='TD')return-1;if(!cell.parentNode||cell.parentNode.tagName!='TR')return-1;var tr=cell.parentNode;var trs=tr.parentNode.childNodes;for(var i=0;i<trs.length;i++)
if(trs[i]==tr)
return i;return-1;},rowID:function(cell)
{if(cell.nodeType!=1)return-1;if(cell.tagName!='TD')return-1;if(!cell.parentNode||cell.parentNode.tagName!='TR')return-1;var tr=cell.parentNode;return tr.id;},sizeTo:function(changeMe,model)
{changeMe.style.position='absolute';changeMe.style.zindex=99;changeMe.style.left=model.offsetLeft+'px';changeMe.style.top=model.offsetTop+'px';changeMe.style.width=model.offsetWidth+'px';changeMe.style.height=model.offsetHeight+'px';return changeMe;},recursiveAddOnClickHandler:function(element)
{element.onclick=InlineEditor.handleOnClick;if(element.childNodes){children=element.childNodes;for(i=0;i<children.length;i++){if(children[i].onclick){InlineEditor.recursiveAddOnClickHandler(children[i]);}}}},handleOnClick:function(evt)
{var evt=InlineEditor.fixEvent(evt);var target=InlineEditor.findEditableTarget(evt.target);if(InlineEditor.checkClass(target,'uneditable')||InlineEditor.checkClass(target,'editing'))
return;var oldHTML=target.innerHTML;var oldVal=null;if(InlineEditor.elementValue)
oldVal=InlineEditor.elementValue(target);if(!oldVal)
oldVal=target.innerHTML;var editor=null;if(InlineEditor.customEditor){editor=InlineEditor.customEditor(target);}
if(!editor){if(target.offsetHeight>20&&target.innerHTML.length>20){editor=document.createElement('textarea');editor.innerHTML=oldVal;editor.style.width=target.offsetWidth+'px';editor.style.height=target.offsetHeight+'px';}
else{editor=document.createElement('input');editor.value=oldVal;editor.style.width=target.offsetWidth+'px';}}
editor.onblur=function(){InlineEditor.handleInputBlur(editor,oldVal,oldHTML);}
InlineEditor.addClass(target,'editing');target.innerHTML="";target.appendChild(editor);editor.focus();return false;},handleInputBlur:function(editor,oldVal,oldHTML)
{var parent=editor.parentNode;var newVal=null;if(InlineEditor.editorValue)
newVal=InlineEditor.editorValue(editor);if(!newVal)
newVal=editor.value?editor.value:editor.innerHTML;if(oldVal==newVal){parent.innerHTML=oldHTML
InlineEditor.removeClass(parent,'editing');return;}
parent.innerHTML=newVal;InlineEditor.removeClass(parent,'editing');if(InlineEditor.elementChanged)
InlineEditor.elementChanged(parent,oldVal,newVal);},jscss:function(a,o,c1,c2)
{switch(a){case'swap':o.className=!InlineEditor.jscss('check',o,c1)?o.className.replace(c2,c1):o.className.replace(c1,c2);break;case'add':if(!InlineEditor.jscss('check',o,c1)){o.className+=o.className?' '+c1:c1;}
break;case'remove':var rep=o.className.match(' '+c1)?' '+c1:c1;o.className=o.className.replace(rep,'');break;case'check':return new RegExp('\\b'+c1+'\\b').test(o.className)
break;}},fixEvent:function(evt)
{var E=evt?evt:window.event;if(E.target)
if(E.target.nodeType==3)
E.target=E.target.parentNode;if(inlineeditor_isIE)
if(E.srcElement)
E.target=E.srcElement;return E;},findEditableTarget:function(target)
{if(target.nodeType==1&&target.tagName=='TD')
return target;if(InlineEditor.checkClass(target,'editable'))
return target;if(target.parentNode)
return InlineEditor.findEditableTarget(target.parentNode);return null;},addEvent:function(target,eventName,func,capture)
{if(target.addEventListener){target.addEventListener(eventName,func,capture);return true;}
else if(target.attachEvent)
return target.attachEvent('on'+eventName,func);},removeEvent:function(target,eventName,func,capture)
{if(target.removeEventListener){target.removeEventListener(eventName,func,capture);return true;}
else if(target.detachEvent)
return target.detachEvent('on'+eventName,func);}}
InlineEditor.addEvent(window,'load',InlineEditor.init,false);