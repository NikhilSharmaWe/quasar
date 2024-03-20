
window.addEventListener("DOMContentLoaded", (_) => { 
  // const {PythonShell} =require('python-shell');

    path = window.location.pathname;

    key = path.replace("/meets/",'').replace("/",'');
    let mk = document.getElementById("meeting_key");
    mk.innerHTML = `<p><strong>Meeting ID: </strong>${key}</p>`;

    // textarea.addEventListener('input', function() {
    //   console.log("HELOO")
    //   const content = textarea.value;
    //   sendUpdate(content);
    // });
    
    // function sendUpdate(content) {
    //   // const message = {
    //   //   type: 'update',
    //   //   content: content
    //   // };
    //   // ws.send(JSON.stringify(message));
    //   ws.send(JSON.stringify({ event: 'code', data: JSON.stringify(content) }));
    // }

    function removeFirstLast(str) { 
      return str.slice(1, -1); 
    } 

    function onTestChange() {
      var key = window.event.keyCode;
  
      // If the user has pressed enter
      if (key === 13) {
          document.getElementById("txtArea").value = document.getElementById("txtArea").value + "\n*";
          return false;
      }
      else {
          return true;
      }
  }

    navigator.mediaDevices.getUserMedia({ video: true, audio: true })
    .then(stream => {

      console.log("HEY")

      const chatInput = document.getElementById('chat-input');
      const chatButton = document.getElementById('send-button');
      const chatMessages = document.getElementById('chat-messages');
      const codeOutput = document.getElementById('code-output');

      chatButton.addEventListener('click', () => {
        console.log("FFFFF")
        const message = chatInput.value;
        if (message.trim() !== '') {
          ws.send(JSON.stringify({ event: 'chat', data: message }));
          chatInput.value = '';
        }
      });

      const textarea = document.getElementById('textarea');
      textarea.addEventListener('input', function() {
        console.log("HELOO")
        const content = textarea.value;
        
        // content = content;
        // sendUpdate(content);
        ws.send(JSON.stringify({ event: 'code', data: JSON.stringify(content) }));

      });

      document.getElementById("run").addEventListener('click', function() {
        const content = textarea.value;
        console.log("COMPILE", content)
        
        // content = content;
        // sendUpdate(content);
        ws.send(JSON.stringify({ event: 'compile', data: JSON.stringify(content) }));
      });
  

      // $(".textarea").keypress(function(event) {
      //   if (event.which == 13) {        
      //        alert("Function is Called on Enter");
      //     }
 
      // });


      const config = {
        codec: 'vp8',
        iceServers: [
            {
                "urls": "stun:stun.l.google.com:19302",
            },
        ]
      }
      let pc = new RTCPeerConnection(config)

      pc.ontrack = function (event) {
        if (event.track.kind === 'audio' || event.track.kind === 'video') {
            let streamID = event.streams[0].id;
            const pElement = document.createElement('span');
            pElement.style.color = 'white';
            pElement.id = event.streams[0].id; // Use the streamID as the id
            document.getElementById('remoteVideos').appendChild(pElement);

            let el = document.createElement(event.track.kind);
            el.srcObject = event.streams[0];
            el.autoplay = true;
            el.controls = true;
            pElement.appendChild(el);

            event.track.onmute = function(event) {
                el.play();
            }

            event.streams[0].onremovetrack = ({track}) => {
              if (el.parentNode) {
                el.parentNode.removeChild(el);
              }
              let e = document.getElementById(streamID);
              e.remove();
            }
        }
      }

      pc.onremovestream = function (event) {
        // console.log("removing")
        const streamId = event.stream.id;
        const element = remoteVideoElements[streamId];
        if (element) {
          element.remove();
        }
      }

      

      document.getElementById('localVideo').srcObject = stream
      stream.getTracks().forEach(track => pc.addTrack(track, stream))
      // console.log("local video data:", stream);
      let ws = new WebSocket("ws://" + window.location.host + "/websocket");
      pc.onicecandidate = e => {
        if (!e.candidate) {
          return
        }

        ws.send(JSON.stringify({event: 'candidate', data: JSON.stringify(e.candidate)}))
      }

      ws.onclose = function(evt) {
        // console.log("websocket connection closed")
        window.alert("Websocket has closed")
      }

      ws.onmessage = function(evt) {
        let msg = JSON.parse(evt.data)
        if (!msg) {
          return console.log('failed to parse msg')
        }

        console.log(msg)

        // console.log(msg);
        // console.log(participantInfo);

        switch (msg.event) {
          case 'offer':
            let offer = JSON.parse(msg.data)
            if (!offer) {
              return console.log('failed to parse answer')
            }
            // console.log("Recived offer")
            pc.setRemoteDescription(offer)
            pc.createAnswer().then(answer => {
              pc.setLocalDescription(answer)
              ws.send(JSON.stringify({event: 'answer', data: JSON.stringify(answer)}))
            })
            return

          case 'candidate':
            let candidate = JSON.parse(msg.data)
            if (!candidate) {
              return console.log('failed to parse candidate')
            }

            pc.addIceCandidate(candidate)
            return
          
          case 'chat':
            const chatData = msg.data;
            const message = chatData.Message;
            const username = chatData.Username;

            // console.log(username + ":" + message)

            const messageElement = document.createElement('div');
            messageElement.innerHTML = `<p><strong>${username}:</strong> ${message}</p>`;
            chatMessages.appendChild(messageElement);
            return
          case 'code':
            console.log("XXXXXXXXXXXXXXXXXXXXXXXXX")
            const codeData = msg.data;
            const code = codeData.Code;
            console.log(code)

            // console.log(username + ":" + message)
            var newline = String.fromCharCode(13, 10);
            const x = removeFirstLast(code).replaceAll('\\n', newline);
            const trimmedCode = x.replaceAll('\\', '');
            // trimmedCode = x.replace('\\', '');
            textarea.value = `${trimmedCode}`;
            return
          
          case 'output':
            const outputData = msg.data;
            console.log("OUTPUT: ", outputData)
            codeOutput.innerHTML = `<p>${outputData}</p>`;




//             const jsonString = JSON.stringify(msg);
//             const parsedData = JSON.parse(jsonString);

// // Extract the value of the Code property
//             const codeValue = parsedData.data.Code;
//             console.log("HERE: ", codeValue)
//             textarea.value = `${codeValue}`;

          // case 'my_username':
          //   let un = document.getElementById("your_username");
          //   un.innerHTML = msg.data;
          //   return

          case 'participant':
            const participantData = msg.data;
            // participantInfo[participantData.StreamID] = participantData.Username;
            let p = document.getElementById(participantData.StreamID);
            p.innerHTML = participantData.Username;
            // console.log(participantInfo);
            return
        }
      }

      ws.onerror = function(evt) {
        console.log("ERROR: " + evt.data)
      }
    }).catch(window.alert)
});
