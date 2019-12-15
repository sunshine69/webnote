/**
 * AJAX: Some simple, handy scripts for asynchronous communications.
 *
 * Version: 0.1
 *
 * License: Public Domain
 *
 * Author: Robert Harder
 *         rharder@users.sf.net
 *
 * Functions for you to call:
 *
 *     AJAX.getXML( url, callback )
 *         Retrieves 'url' and attempts to parse as XML.
 *         The xml DOM object is passed to the 'callback'
 *         function which should accept a single argument,
 *         the xml DOM object.
 *
 *     AJAX.getText( url, callback )
 *         Retrieves 'url' and passes the raw returned text
 *         to the 'callback' function which should accept a
 *         single argument, the text string.
 *
 *     AJAX.setValue( url, element )
 *         Retrieves raw text from 'url' and attempts to set
 *         the 'value' property of 'element'.
 *
 *     AJAX.setValueById( url, id )
 *         Retrieves raw text from 'url' and attempts to set
 *         the 'value' property of element 'id'.
 *
 *     AJAX.setInnerHTML( url, element )
 *         Retrieves raw text from 'url' and attempts to set
 *         the 'innerHTML' property of 'element'.
 *
 *     AJAX.setInnerHTMLById( url, id )
 *         Retrieves raw text from 'url' and attempts to set
 *         the 'innerHTML' property of element 'id'.
 *
 *
 * Condensed Example:
 *
 *     <p>
 *       The answer to Life, the Universe, and Everything:
 *       <span id="answer">Waiting for Deep Thought...</span>
 *     </p>
 *     <script src="ajax.js" type="text/javascript" ></script>
 *     <script>
 *       AJAX.setInnerHTMLById( 'deepthought.php?action=thinkhard', answer )
 *     </script>
 *
 */
//Update by stevek to add some small functions I used.
 ////
// If AJAX is not yet defined, define it.
// This protects against AJAX accidentally
// being included more than once.
////
if( typeof AJAX == 'undefined' ){

    function trim(s) { return  s.replace(/^\s+|\s+$/g, '') ;}

    function myEncode(str) {
        var s=escape(trim(str));
        s=s.replace(/\+/g,"+");
        s=s.replace(/@/g,"@");
        s=s.replace(/\//g,"/");
        s=s.replace(/\*/g,"*");
        return(s);
    }

    function copyTextToClipboard(text_to_copy) {
      var textArea = document.createElement("textarea");
      textArea.style.position = 'fixed';
      textArea.style.top = 0;
      textArea.style.left = 0;
      textArea.style.width = '2em';
      textArea.style.height = '2em';
      textArea.style.padding = 0;
      textArea.style.border = 'none';
      textArea.style.outline = 'none';
      textArea.style.boxShadow = 'none';
      textArea.style.background = 'transparent';

      textArea.value = text_to_copy;
      document.body.appendChild(textArea);

      textArea.select();

      try {
        var successful = document.execCommand('copy');
        var msg = successful ? 'successful' : 'unsuccessful';
        console.log('Copying text command was ' + msg);
      } catch (err) {
        console.log('Oops, unable to copy');
      }

      document.body.removeChild(textArea);
    }

    // Some functions to generate random password
    String.prototype.pick = function(min, max) {
        var n, chars = '';

        if (typeof max === 'undefined') {
            n = min;
        } else {
            n = min + Math.floor(Math.random() * (max - min + 1));
        }

        for (var i = 0; i < n; i++) {
            chars += this.charAt(Math.floor(Math.random() * this.length));
        }

        return chars;
    };
    // Credit to @Christoph: http://stackoverflow.com/a/962890/464744
    String.prototype.shuffle = function() {
        var array = this.split('');
        var tmp, current, top = array.length;

        if (top) while (--top) {
            current = Math.floor(Math.random() * (top + 1));
            tmp = array[current];
            array[current] = array[top];
            array[top] = tmp;
        }

        return array.join('');
    };

    function pwgen(length) {
        var specials = '!@#%.,<>{}$^&-_=+';
        var lowercase = 'abcdefghijklmnopqrstuvwxyz';
        var uppercase = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ';
        var numbers = '0123456789';

        var all = specials + lowercase + uppercase + numbers;

        var password = '';
        password += specials.pick(1);
        password += lowercase.pick(1);
        password += uppercase.pick(1);
        password += numbers.pick(1);
        password += all.pick(length, length);
        password = password.shuffle();
        return password;
    }

    AJAX = {
        /**
         * Retrieves 'url' and attempts to parse as XML.
         * The xml DOM object is passed to the 'callback'
         * function which should accept a single argument,
         * the xml DOM object.
         */
        getXML : function( url, callback )
        {
            return AJAX.ajaxFull( url, callback, false );
        },  // end getXML

        /**
         * Retrieves 'url' and passes the raw returned text
         * to the 'callback' function which should accept a
         * single argument, the text string.
         */
        getText : function( url, callback )
        {
            return AJAX.ajaxFull( url, callback, true );
        },  // end getText

        postText : function( url, params, callback)
        { return AJAX.ajaxPost( url, callback, params, true ); },

        postTextBasicAuth : function( url, params, user, pass, callback)
        { return AJAX.ajaxPost( url, callback, params, true, user, pass ); },

        /**
         * Retrieves raw text from 'url' and attempts to set
         * the 'innerHTML' property of 'element'.
         */
        setInnerHTML: function ( url, element )
        {
            AJAX.getText( url, function( text ){

                if( element && element.innerHTML )
                    element.innerHTML = text;

            }); // end ajax function
        },  // end setInnerHTML

        /**
         * Retrieves raw text from 'url' and attempts to set
         * the 'innerHTML' property of element 'id'.
         */
        setInnerHTMLById : function( url, id )
        {
            if( document.getElementById )
                return AJAX.setInnerHTML( url, document.getElementById( id ) );

        },  // end setInnerHTMLById

        /**
         * Retrieves raw text from 'url' and attempts to set
         * the 'value' property of 'element'.
         */
        setValue: function ( url, element )
        {
            AJAX.getText( url, function( text ){

                if( element && element.value )
                    element.value = text;

            }); // end ajax function
        },  // end setInnerHTML

        /**
         * Retrieves raw text from 'url' and attempts to set
         * the 'value' property of element 'id'.
         */
        setValueById : function( url, id )
        {
            if( document.getElementById )
                return AJAX.setValue( url, document.getElementById( id ) );

        },  // end setValueById

    /* ********  I N T E R N A L   F U N C T I O N S  ******** */
        ajaxPost : function( url, callback, params, textInsteadOfXml, user='', pass='') {
        var request = AJAX.httprequest();
        if (user && pass) request.open("POST", url, true, user, pass);
        else request.open("POST", url, true);
        request.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
        // This caused error: Refused to set unsafe header
        //request.setRequestHeader("Content-length", params.length);
        //request.setRequestHeader("connection", "close");
        request.send(params);
        request.onreadystatechange = function() {
        if( request.readyState == 4 ) {
            // Text
            if( textInsteadOfXml ) {
                callback( request.responseText );
            } // end if: text
            // XML
            else {
                var xmlDoc = request.responseXML;
                // Special case: if we're using Google's stuff
                // use their parser as a fallback.
                if( xmlDoc.documentElement == null && GXml && GXml.parse )
                xmlDoc = GXml.parse( request.responseText );
                callback( xmlDoc );
            } // end else: xml
            } // end if: ready state 4
        }; // end on ready state change

        },

        ////
        // Used internally to retrieve text asynchronously.
        ////
        ajaxFull : function( url, callback, textInsteadOfXml )
        {
            var request = AJAX.httprequest();
            request.open("GET", url, true);
            request.onreadystatechange = function() {
                if( request.readyState == 4 ) {

                    // Text
                    if( textInsteadOfXml ) {

                        callback( request.responseText );

                    } // end if: text

                    // XML
                    else {

                        var xmlDoc = request.responseXML;

                        // Special case: if we're using Google's stuff
                        // use their parser as a fallback.
                        if( xmlDoc.documentElement == null && GXml && GXml.parse )
                            xmlDoc = GXml.parse( request.responseText );

                        callback( xmlDoc );

                    } // end else: xml
                } // end if: ready state 4
            }; // end on ready state change

            request.send(null);

        },// end ajax


        ////
        // Used internally to create HttpRequest.
        ////
        httprequest : function()
        {
            // Microsoft?
            if( typeof ActiveXObject != 'undefined' ){
                try {
                    return new ActiveXObject( 'Microsoft.XMLHTTP' );
                } catch( exc ) {
                    // error
                }   // end catch: exception
            }   // end if: Microsoft

            // Standard?
            if( typeof XMLHttpRequest != 'undefined' ){
                try {
                    return new XMLHttpRequest();
                } catch( exc ) {
                    // error
                }   // end catch: exception
            }   // end if: Standard
        }

    }   // end AJAX
    }   // end if: AJAX not already defined