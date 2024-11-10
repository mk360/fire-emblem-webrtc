let localStream;
let peerConnection;
let isMuted = false;
let isVideoStopped = false;

document.onload = async function() {
    peerConnection = new RTCPeerConnection({
        iceServers: [{
            urls: "stun:stun.l.google.com:19302"
        }]
    });

    const ws = new WebSocket("ws://http://localhost:8080/ws");
    ws.send(JSON.stringify({ type: "test" }));

    ws.onmessage = async (message) => {

        const data = JSON.parse(message.data);

        switch (data.type) {

            case 'offer':
                await peerConnection.setRemoteDescription(new RTCSessionDescription(data.offer));
                const answer = await peerConnection.createAnswer();
                await peerConnection.setLocalDescription(answer);
                ws.send(JSON.stringify({ type: 'answer', answer: answer }));
                break;

            case 'answer':
                await peerConnection.setRemoteDescription(new RTCSessionDescription(data.answer));
                break;

            case 'candidate':
                await peerConnection.addIceCandidate(new RTCIceCandidate(data.candidate));
                break;

            default:
                break;

        }
    };

    peerConnection.onicecandidate = (event) => {
        if (event.candidate) {
            ws.send(JSON.stringify({ type: 'candidate', candidate: event.candidate }));
        }

    };


    peerConnection.ontrack = (event) => {
        addRemoteStream(event.streams[0]);
    };

    peerConnection.onremovetrack = (event) => {
        const videoElements = document.getElementById('videos').getElementsByTagName('video');
    
        for (let video of videoElements) {
            if (video.srcObject === event.streams[0]) {
                video.remove();
                break;
            }
        }
    };
}

function addRemoteStream(stream) {

    const remoteVideo = document.createElement('video');

    remoteVideo.srcObject = stream;

    remoteVideo.autoplay = true;

    document.getElementById('videos').appendChild(remoteVideo);

}
