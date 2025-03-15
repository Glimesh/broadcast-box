// Changes to the React Chat component

import { useState, useEffect, useRef } from 'react';
import './styles.css';

const Chat = ({ roomName = 'default', username: usernameProp = null }) => {
  // Generate random username once on component mount
  const [username, setUsername] = useState(() => {
    return usernameProp || `user_${Math.floor(Math.random() * 1000)}`;
  });
  
  const [messages, setMessages] = useState([]);
  const [newMessage, setNewMessage] = useState('');
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState(null);
  const [userCount, setUserCount] = useState(0);
  
  const webSocketRef = useRef(null);
  const chatContainerRef = useRef(null);
  const hasConnectedRef = useRef(false);
  
  // Auto scroll to bottom when messages change
  useEffect(() => {
    if (chatContainerRef.current) {
      chatContainerRef.current.scrollTop = chatContainerRef.current.scrollHeight;
    }
  }, [messages]);
  
  // Connect to WebSocket on component mount
  useEffect(() => {
    // Only connect if we haven't already
    if (!hasConnectedRef.current) {
      connectToChat();
      hasConnectedRef.current = true;
    }
    
    // Clean up on unmount
    return () => {
      disconnectFromChat();
    };
  }, [roomName]);
  
  // Handle page unload/refresh
  useEffect(() => {
    const handleUnload = () => {
      disconnectFromChat();
    };
    
    window.addEventListener('beforeunload', handleUnload);
    
    return () => {
      window.removeEventListener('beforeunload', handleUnload);
    };
  }, []);
  
  const connectToChat = () => {
    // Don't create a new connection if one already exists
    if (webSocketRef.current) {
      console.log('WebSocket connection already exists');
      return;
    }
    
    // Determine WebSocket protocol (ws or wss)
    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const host = window.location.host;
    
    // Get API path from environment variables, with fallback to deprecated REACT_APP_API_PATH
    // Using the same approach as the player component for consistency
    const apiPath = import.meta.env.VITE_API_PATH ?? (() => {
      console.warn('[broadcast box] REACT_APP_API_PATH is deprecated, please use VITE_API_PATH instead');
      return import.meta.env.REACT_APP_API_PATH;
    })();
    
    // Build WebSocket URL with username parameter - using same logic as player component
    let wsUrl;
    if (apiPath && (apiPath.startsWith('http://') || apiPath.startsWith('https://'))) {
      // Use the full URL from environment
      wsUrl = `${protocol}//${host}${apiPath}/chat?room=${encodeURIComponent(roomName)}&username=${encodeURIComponent(username)}`;
    } else if (apiPath) {
      // It's just a path, use with current host
      wsUrl = `${protocol}//${host}${apiPath}/chat?room=${encodeURIComponent(roomName)}&username=${encodeURIComponent(username)}`;
    } else {
      // No API path, just use current host with /api prefix
      wsUrl = `${protocol}//${host}/api/chat?room=${encodeURIComponent(roomName)}&username=${encodeURIComponent(username)}`;
    }
    
    // Fix the URL if it has a double path issue - there shouldn't be /api/api/
    wsUrl = wsUrl.replace('/api/api/', '/api/');
    // Fix the endpoint if necessary - ensure we use /api/chat
    if (!wsUrl.includes('/api/chat')) {
      wsUrl = wsUrl.replace('/chat', '/api/chat');
    }
    
    console.log('Connecting to WebSocket URL:', wsUrl);
    
    try {
      webSocketRef.current = new WebSocket(wsUrl);
      
      webSocketRef.current.onopen = () => {
        console.log('WebSocket connection established');
        setIsConnected(true);
        setError(null);
        
        // Add system message
        addSystemMessage(`Connected to chat room: ${roomName}`);
        
        // Send join message
        sendMessage('join', `joined the chat`);
      };
      
      webSocketRef.current.onmessage = (event) => {
        try {
          const message = JSON.parse(event.data);
          
          if (message.type === 'message') {
            // Only process messages from others
            // We already added our own messages locally in handleSendMessage
            if (message.sender !== username) {
              // This is a message from someone else, so add it
              addChatMessage(message.sender || 'Anonymous', message.content, false);
            }
          } else if (message.type === 'join' || message.type === 'leave') {
            addSystemMessage(`${message.sender || 'Someone'} ${message.content}`);
          } else if (message.type === 'usercount') {
            // Update user count
            const count = parseInt(message.content, 10);
            console.log(`Received user count: ${count}`);
            setUserCount(count);
          } else if (message.type === 'system') {
            // System messages (like nickname changes)
            addSystemMessage(message.content);
          }
        } catch (err) {
          console.error('Error parsing message:', err);
          addSystemMessage('Received malformed message');
        }
      };
      
      webSocketRef.current.onclose = (event) => {
        console.log(`WebSocket closed with code: ${event.code}`);
        setIsConnected(false);
        addSystemMessage('Disconnected from chat server');
        webSocketRef.current = null;
      };
      
      webSocketRef.current.onerror = (err) => {
        console.error('WebSocket error:', err);
        setError('Connection error. Please try again later.');
        setIsConnected(false);
      };
      
    } catch (err) {
      console.error('Failed to connect:', err);
      setError('Failed to connect to chat server');
      setIsConnected(false);
      webSocketRef.current = null;
    }
  };
  
  const disconnectFromChat = () => {
    if (webSocketRef.current) {
      try {
        if (webSocketRef.current.readyState === WebSocket.OPEN) {
          // Send leave message
          sendMessage('leave', 'left the chat');
          // Close the connection
          webSocketRef.current.close();
        }
      } catch (err) {
        console.error('Error during disconnect:', err);
      } finally {
        webSocketRef.current = null;
      }
    }
  };
  
  const sendMessage = (type, content, additionalFields = {}) => {
    if (!webSocketRef.current || webSocketRef.current.readyState !== WebSocket.OPEN) {
      return;
    }
    
    const message = {
      type,
      sender: username,
      content,
      ...additionalFields
    };
    
    try {
      webSocketRef.current.send(JSON.stringify(message));
    } catch (err) {
      console.error('Error sending message:', err);
      addSystemMessage('Failed to send message');
    }
  };
  
  const handleSendMessage = (e) => {
    e.preventDefault();
    
    if (!newMessage.trim() || !isConnected) {
      return;
    }
    
    const trimmedMessage = newMessage.trim();
    
    // Check if message is a nickname change command
    if (trimmedMessage.startsWith('/nick ')) {
      const newNickname = trimmedMessage.substring(6).trim();
      
      if (newNickname) {
        // Send nickname change message
        sendMessage('nickname', '', { newNick: newNickname });
        
        // Update local username state
        setUsername(newNickname);
      } else {
        // Add system message if nickname is empty
        addSystemMessage('Please provide a valid nickname, e.g., /nick YourNewName');
      }
    } else {
      // Generate a unique message ID
      const msgId = Date.now().toString() + Math.random().toString().substring(2, 8);
      
      // Add message locally first with the unique ID
      const timestamp = new Date().toLocaleTimeString();
      const newMsg = {
        id: msgId,
        sender: username,
        content: trimmedMessage,
        timestamp,
        type: 'self'
      };
      
      // Add directly to state to avoid duplication
      setMessages(prevMessages => [...prevMessages, newMsg]);
      
      // Send message with ID to allow deduplication
      sendMessage('message', trimmedMessage, { messageId: msgId });
    }
    
    // Clear input
    setNewMessage('');
  };
  
  const addChatMessage = (sender, content, isSelf) => {
    const timestamp = new Date().toLocaleTimeString();
    
    setMessages(prevMessages => [
      ...prevMessages,
      {
        id: Date.now(),
        sender,
        content,
        timestamp,
        type: isSelf ? 'self' : 'other'
      }
    ]);
  };
  
  const addSystemMessage = (content) => {
    const timestamp = new Date().toLocaleTimeString();
    
    setMessages(prevMessages => [
      ...prevMessages,
      {
        id: Date.now(),
        sender: 'System',
        content,
        timestamp,
        type: 'system'
      }
    ]);
  };
  
  const reconnect = () => {
    disconnectFromChat();
    setTimeout(() => {
      hasConnectedRef.current = false;
      connectToChat();
    }, 1000);
  };
  
  return (
    <div className="chat-component">
      <div className="chat-header">
        <h3>Chat: {roomName}</h3>
        <div className="chat-info">
          <div className="user-info">
            <span className="current-user">Your nickname: {username}</span>
            <span className="nickname-hint">(Change with /nick command)</span>
          </div>
          <div className="user-count">
            <span className="user-count-label">Users online:</span>
            <span className="user-count-number">{userCount}</span>
          </div>
          <div className="chat-status">
            Status: {isConnected ? <span className="status-connected">Connected</span> : <span className="status-disconnected">Disconnected</span>}
            {!isConnected && (
              <button onClick={reconnect} className="reconnect-button">
                Reconnect
              </button>
            )}
          </div>
        </div>
      </div>
      
      {error && <div className="chat-error">{error}</div>}
      
      <div className="chat-messages" ref={chatContainerRef}>
        {messages.length === 0 ? (
          <div className="chat-empty">No messages yet</div>
        ) : (
          messages.map(message => (
            <div key={message.id} className={`chat-message ${message.type}`}>
              <div className="message-header">
                <span className="message-sender">{message.sender}</span>
                <span className="message-time">{message.timestamp}</span>
              </div>
              <div className="message-content">{message.content}</div>
            </div>
          ))
        )}
      </div>
      
      <form onSubmit={handleSendMessage} className="chat-input-form">
        <input
          type="text"
          value={newMessage}
          onChange={(e) => setNewMessage(e.target.value)}
          placeholder="Type a message or /nick YourNewName"
          disabled={!isConnected}
          className="chat-input"
        />
        <button 
          type="submit" 
          disabled={!isConnected || !newMessage.trim()}
          className="chat-send-button"
        >
          Send
        </button>
      </form>
    </div>
  );
};

export default Chat;