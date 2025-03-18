// Changes to the React Chat component

import { useState, useEffect, useRef, useCallback } from 'react';
import './styles.css';

const Chat = ({ roomName = 'default', username: usernameProp = null }) => {
  // Get saved nickname from localStorage or generate a random one
  const [username, setUsername] = useState(() => {
    // Try to get saved username from localStorage first
    const savedUsername = localStorage.getItem(`broadcast-box-username`);
    
    // Use in order: prop, saved username, or random name
    return usernameProp || savedUsername || `user_${Math.floor(Math.random() * 1000)}`;
  });
  
  // Load messages from localStorage or start with empty array
  const [messages, setMessages] = useState(() => {
    try {
      const savedMessages = localStorage.getItem(`broadcast-box-messages-${roomName}`);
      
      // If we have saved messages, mark them as from history
      if (savedMessages) {
        const parsedMessages = JSON.parse(savedMessages);
        // Add fromHistory flag to differentiate saved messages
        return parsedMessages.map(msg => ({
          ...msg,
          fromHistory: true
        }));
      }
      return [];
    } catch (err) {
      console.error('Failed to load saved messages:', err);
      return [];
    }
  });
  
  const [newMessage, setNewMessage] = useState('');
  const [isConnected, setIsConnected] = useState(false);
  const [error, setError] = useState(null);
  const [userCount, setUserCount] = useState(0);
  const [lastPingTime, setLastPingTime] = useState(null);
  const [userList, setUserList] = useState([]);
  const [userListOpen, setUserListOpen] = useState(false);
  
  const webSocketRef = useRef(null);
  const chatContainerRef = useRef(null);
  const hasConnectedRef = useRef(false);
  
  // Function to save messages to localStorage
  const saveMessagesToStorage = useCallback(() => {
    try {
      // Only keep latest 200 messages to prevent exceeding storage limits
      const messagesToSave = messages.slice(-200);
      localStorage.setItem(`broadcast-box-messages-${roomName}`, JSON.stringify(messagesToSave));
    } catch (err) {
      console.error('Failed to save messages to localStorage:', err);
    }
  }, [messages, roomName]);
  
  // Auto scroll to bottom when messages change and save to localStorage
  useEffect(() => {
    if (chatContainerRef.current) {
      chatContainerRef.current.scrollTop = chatContainerRef.current.scrollHeight;
    }
    
    // Save messages to localStorage
    saveMessagesToStorage();
  }, [messages, saveMessagesToStorage]);
  
  // Connect to WebSocket on component mount
  useEffect(() => {
    // Only connect if we haven't already
    if (!hasConnectedRef.current) {
      connectToChat();
      hasConnectedRef.current = true;
    }
    
    // Set up a heartbeat interval to keep connection alive
    const heartbeatInterval = setInterval(() => {
      if (webSocketRef.current && webSocketRef.current.readyState === WebSocket.OPEN) {
        console.log('Sending client ping to server');
        try {
          webSocketRef.current.send(JSON.stringify({ type: 'ping' }));
          
          // Update timestamp even when sending a ping
          setLastPingTime(new Date());
        } catch (err) {
          console.error('Error sending ping:', err);
        }
      }
    }, 60000); // Send a ping every minute
    
    // Set up a connection status check interval
    const connectionCheckInterval = setInterval(() => {
      // If we're connected but haven't received a ping in 35 minutes (a bit more than pongWait), 
      // the connection is probably stale despite WebSocket status showing OPEN
      if (isConnected && lastPingTime) {
        const timeSinceLastPing = new Date().getTime() - lastPingTime.getTime();
        if (timeSinceLastPing > 35 * 60 * 1000) {
          console.log('Connection appears stale, reconnecting...');
          reconnect();
        }
      }
    }, 60000); // Check every minute
    
    // Set initial ping time when connecting
    setLastPingTime(new Date());
    
    // Clean up on unmount
    return () => {
      console.log('Cleaning up chat resources...');
      clearInterval(heartbeatInterval);
      clearInterval(connectionCheckInterval);
      disconnectFromChat();
      // Ensure refs are cleared
      webSocketRef.current = null;
      hasConnectedRef.current = false;
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
    
    console.log('Starting chat connection process for room:', roomName);
    
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
      // Extract the host from the full URL (without protocol)
      const apiHost = apiPath.replace(/^https?:\/\//, '');
      // Use the extracted host with WebSocket protocol
      wsUrl = `${protocol}//${apiHost}/api/chat?room=${encodeURIComponent(roomName)}&username=${encodeURIComponent(username)}`;
    } else if (apiPath) {
      // It's just a path, use with current host
      // Make sure we don't have a leading slash in apiPath when appending to avoid double slashes
      const formattedApiPath = apiPath.startsWith('/') ? apiPath : `/${apiPath}`;
      wsUrl = `${protocol}//${host}${formattedApiPath}/chat?room=${encodeURIComponent(roomName)}&username=${encodeURIComponent(username)}`;
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
    
    // Trim any trailing slashes before the ?
    wsUrl = wsUrl.replace(/\/\?/, '?');
    
    console.log('Connecting to WebSocket URL:', wsUrl);
    
    try {
      webSocketRef.current = new WebSocket(wsUrl);
      
      webSocketRef.current.onopen = () => {
        console.log('WebSocket connection established');
        setIsConnected(true);
        setError(null);
        
        // Send join message without system notification
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
            // Don't add system messages for join/leave anymore
            // User list will show connection status instead
          } else if (message.type === 'usercount') {
            // Update user count
            const count = parseInt(message.content, 10);
            console.log(`Received user count: ${count}`);
            setUserCount(count);
          } else if (message.type === 'userlist') {
            // Update user list
            console.log('Received user list:', message.users);
            setUserList(message.users || []);
          } else if (message.type === 'system') {
            // System messages (like nickname changes)
            addSystemMessage(message.content);
          } else if (message.type === 'ping') {
            // Server sent a ping, respond with a pong
            console.log('Received ping from server');
            if (webSocketRef.current && webSocketRef.current.readyState === WebSocket.OPEN) {
              webSocketRef.current.send(JSON.stringify({ type: 'pong' }));
            }
            // Update last ping time
            setLastPingTime(new Date());
          } else if (message.type === 'pong') {
            // Server sent a pong response
            console.log('Received pong from server');
            // Update last ping time on pong too
            setLastPingTime(new Date());
          }
        } catch (err) {
          console.error('Error parsing message:', err);
          addSystemMessage('Received malformed message');
        }
      };
      
      webSocketRef.current.onclose = (event) => {
        console.log(`WebSocket closed with code: ${event.code}`);
        setIsConnected(false);
        webSocketRef.current = null;
        // Clear the user list when disconnected
        setUserList([]);
      };
      
      webSocketRef.current.onerror = (err) => {
        console.error('WebSocket error:', err);
        setError('Connection error. Please try again later.');
        setIsConnected(false);
        
        // Try to reconnect automatically after an error
        setTimeout(() => {
          console.log('Attempting to reconnect after error...');
          if (!isConnected && !webSocketRef.current) {
            reconnect();
          }
        }, 5000);
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
      console.log('Disconnecting from chat...');
      try {
        // First check if we can send a leave message
        if (webSocketRef.current.readyState === WebSocket.OPEN) {
          // Send leave message without system notification
          sendMessage('leave', 'left the chat');
          // Close the connection properly
          webSocketRef.current.close(1000, 'User disconnected');
        } else if (webSocketRef.current.readyState === WebSocket.CONNECTING) {
          // If still connecting, abort the connection
          webSocketRef.current.close(1000, 'Connection aborted');
        }
      } catch (err) {
        console.error('Error during disconnect:', err);
      } finally {
        // Clean up the reference and update connection state
        webSocketRef.current = null;
        setIsConnected(false);
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
        
        // Save the new nickname to localStorage
        localStorage.setItem('broadcast-box-username', newNickname);
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
    
    // Clear input and any saved draft
    setNewMessage('');
    clearDraft();
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
  
  const addSystemMessage = useCallback((content) => {
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
  }, []);
  
  const reconnect = () => {
    console.log('Reconnecting to chat...');
    // Clean up any existing connection
    disconnectFromChat();
    
    // Clear any previous errors
    setError(null);
    
    // Reset the connection state to allow a fresh connection
    hasConnectedRef.current = false;
    webSocketRef.current = null;
    
    // Add a small delay before reconnecting
    setTimeout(() => {
      connectToChat();
    }, 1000);
  };
  
  // Function to clear chat history
  const clearChatHistory = useCallback(() => {
    // Remove messages from state
    setMessages([]);
    
    // Clear localStorage for this room
    localStorage.removeItem(`broadcast-box-messages-${roomName}`);
    
    // Add system message indicating history was cleared
    addSystemMessage('Chat history has been cleared');
  }, [roomName, addSystemMessage]);
  
  // State variable to track if a draft exists
  const [hasSavedDraft, setHasSavedDraft] = useState(false);
  
  // Check for draft existence
  const checkDraftExists = useCallback(() => {
    const exists = Boolean(localStorage.getItem(`broadcast-box-draft-${roomName}`));
    setHasSavedDraft(exists);
    return exists;
  }, [roomName]);
  
  // Helper to check if there are unsaved draft messages in input
  const hasDraft = useCallback(() => {
    return hasSavedDraft;
  }, [hasSavedDraft]);
  
  // Manual draft operations
  const saveDraft = useCallback(() => {
    if (newMessage.trim()) {
      localStorage.setItem(`broadcast-box-draft-${roomName}`, newMessage);
      // Show a system message confirming the draft was saved
      addSystemMessage('Draft message saved');
      // Update draft status
      setHasSavedDraft(true);
    }
  }, [newMessage, roomName]);
  
  const loadDraft = useCallback(() => {
    const savedDraft = localStorage.getItem(`broadcast-box-draft-${roomName}`);
    if (savedDraft) {
      setNewMessage(savedDraft);
      return true;
    }
    return false;
  }, [roomName]);
  
  const clearDraft = useCallback(() => {
    localStorage.removeItem(`broadcast-box-draft-${roomName}`);
    setHasSavedDraft(false);
  }, [roomName]);
  
  // Check for drafts and set up autosave on unmount
  useEffect(() => {
    // Check for drafts on component mount
    checkDraftExists();
    
    // Load draft when component mounts (only once)
    loadDraft();
    
    // Add event listener for page unload to auto-save
    const autoSaveOnUnload = () => {
      if (newMessage.trim()) {
        localStorage.setItem(`broadcast-box-draft-${roomName}`, newMessage);
        setHasSavedDraft(true);
      } else {
        localStorage.removeItem(`broadcast-box-draft-${roomName}`);
        setHasSavedDraft(false);
      }
    };
    
    window.addEventListener('beforeunload', autoSaveOnUnload);
    
    // Clean up event listeners
    return () => {
      autoSaveOnUnload();
      window.removeEventListener('beforeunload', autoSaveOnUnload);
    };
  }, [roomName, loadDraft, checkDraftExists]); // Don't include newMessage to prevent constant updates
  
  return (
    <div className="chat-component">
      <div className="chat-header">
        <div className="chat-header-row">
          <h3>Chat: {roomName}</h3>
          <div className="chat-header-buttons">
            <button 
              onClick={clearChatHistory}
              className="clear-history-button"
              title="Clear chat history"
              aria-label="Clear chat history"
            >
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" width="16" height="16">
                <path fillRule="evenodd" d="M16.5 4.478v.227a48.816 48.816 0 0 1 3.878.512.75.75 0 1 1-.256 1.478l-.209-.035-1.005 13.07a3 3 0 0 1-2.991 2.77H8.084a3 3 0 0 1-2.991-2.77L4.087 6.66l-.209.035a.75.75 0 0 1-.256-1.478A48.567 48.567 0 0 1 7.5 4.705v-.227c0-1.564 1.213-2.9 2.816-2.951a52.662 52.662 0 0 1 3.369 0c1.603.051 2.815 1.387 2.815 2.951Zm-6.136-1.452a51.196 51.196 0 0 1 3.273 0C14.39 3.05 15 3.684 15 4.478v.113a49.488 49.488 0 0 0-6 0v-.113c0-.794.609-1.428 1.364-1.452Zm-.355 5.945a.75.75 0 1 0-1.5.058l.347 9a.75.75 0 1 0 1.499-.058l-.346-9Zm5.48.058a.75.75 0 1 0-1.498-.058l-.347 9a.75.75 0 0 0 1.5.058l.345-9Z" clipRule="evenodd" />
              </svg>
            </button>
            <button 
              onClick={() => setUserListOpen(true)} 
              className="show-users-button"
              disabled={userList.length === 0}
              aria-label="Show user list"
            >
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" width="20" height="20">
                <path fillRule="evenodd" d="M8.25 6.75a3.75 3.75 0 1 1 7.5 0 3.75 3.75 0 0 1-7.5 0ZM15.75 9.75a3 3 0 1 1 6 0 3 3 0 0 1-6 0ZM2.25 9.75a3 3 0 1 1 6 0 3 3 0 0 1-6 0ZM6.31 15.117A6.745 6.745 0 0 1 12 12a6.745 6.745 0 0 1 6.709 7.498.75.75 0 0 1-.372.568A12.696 12.696 0 0 1 12 21.75c-2.305 0-4.47-.612-6.337-1.684a.75.75 0 0 1-.372-.568 6.787 6.787 0 0 1 1.019-4.38Z" clipRule="evenodd" />
                <path d="M5.082 14.254a8.287 8.287 0 0 0-1.308 5.135 9.687 9.687 0 0 1-1.764-.44l-.115-.04a.563.563 0 0 1-.373-.487l-.01-.121a3.75 3.75 0 0 1 3.57-4.047ZM20.226 19.389a8.287 8.287 0 0 0-1.308-5.135 3.75 3.75 0 0 1 3.57 4.047l-.01.121a.563.563 0 0 1-.373.486l-.115.04c-.567.2-1.156.349-1.764.441Z" />
              </svg>
              <span className="users-count-badge">{userCount}</span>
            </button>
          </div>
        </div>
        {!isConnected && (
          <div className="reconnect-container">
            <span className="disconnected-message">Disconnected from chat</span>
            <button onClick={reconnect} className="reconnect-button">
              Reconnect
            </button>
          </div>
        )}
      </div>
      
      {error && <div className="chat-error">{error}</div>}
      
      <div className="chat-container">
        {/* User list drawer overlay */}
        <div className={`user-list-drawer ${userListOpen ? 'open' : 'closed'}`}>
          <div className="user-list-header">
            <span>Users in Chat</span>
            <button 
              onClick={() => setUserListOpen(false)}
              className="close-drawer-button"
              aria-label="Close user list"
            >
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" width="20" height="20">
                <path fillRule="evenodd" d="M5.47 5.47a.75.75 0 0 1 1.06 0L12 10.94l5.47-5.47a.75.75 0 1 1 1.06 1.06L13.06 12l5.47 5.47a.75.75 0 1 1-1.06 1.06L12 13.06l-5.47 5.47a.75.75 0 0 1-1.06-1.06L10.94 12 5.47 6.53a.75.75 0 0 1 0-1.06Z" clipRule="evenodd" />
              </svg>
            </button>
          </div>
          <div className="user-list">
            {userList.length === 0 ? (
              <div className="user-list-empty">No users online</div>
            ) : (
              userList.map((user, index) => (
                <div 
                  key={index} 
                  className="user-list-item"
                  title={user.username === username && lastPingTime ? 
                    `Last heartbeat: ${lastPingTime.toLocaleTimeString()} ${new Date().getTime() - lastPingTime.getTime() < 120000 ? '❤️' : ''}` : 
                    undefined
                  }
                >
                  <span className={`user-status ${user.status}`}></span>
                  <span className="user-name">
                    {user.username}
                    {user.username === username && <span className="current-user-indicator"> (You)</span>}
                  </span>
                </div>
              ))
            )}
          </div>
        </div>
        
        {/* Chat messages container */}
        <div className="chat-messages-container">
          
          <div className="chat-messages" ref={chatContainerRef}>
            {messages.length === 0 ? (
              <div className="chat-empty">No messages yet</div>
            ) : (
              <>
                {/* Add separator between history and new messages if needed */}
                {messages.some(msg => msg.fromHistory) && 
                 messages.some(msg => !msg.fromHistory) && (
                  <div className="history-separator">
                    <span>
                      — Loaded {messages.filter(msg => msg.fromHistory).length} previous messages —
                    </span>
                  </div>
                )}
                
                {/* Show only history messages notice */}
                {messages.every(msg => msg.fromHistory) && messages.length > 0 && (
                  <div className="history-notice">
                    <span>
                      Loaded {messages.length} messages from previous session
                    </span>
                  </div>
                )}
                
                {messages.map(message => (
                  <div 
                    key={message.id} 
                    className={`chat-message ${message.type} ${message.fromHistory ? 'from-history' : ''}`}
                  >
                    <div className="message-header">
                      <span className="message-sender">{message.sender}</span>
                      <span className="message-time">
                        {message.timestamp}
                      </span>
                    </div>
                    <div className="message-content">{message.content}</div>
                  </div>
                ))}
              </>
            )}
          </div>
        </div>
      </div>
      
      <form onSubmit={handleSendMessage} className="chat-input-form">
        <div className="chat-input-container">
          <input
            type="text"
            value={newMessage}
            onChange={(e) => setNewMessage(e.target.value)}
            placeholder={`Type a message as ${username} or /nick YourNewName`}
            disabled={!isConnected}
            className="chat-input"
          />
          {newMessage.trim() && (
            <button 
              type="button"
              onClick={saveDraft}
              className="chat-save-draft-button"
              title="Save as draft"
            >
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" width="16" height="16">
                <path d="M21.731 2.269a2.625 2.625 0 0 0-3.712 0l-1.157 1.157 3.712 3.712 1.157-1.157a2.625 2.625 0 0 0 0-3.712ZM19.513 8.199l-3.712-3.712-8.4 8.4a5.25 5.25 0 0 0-1.32 2.214l-.8 2.685a.75.75 0 0 0 .933.933l2.685-.8a5.25 5.25 0 0 0 2.214-1.32l8.4-8.4Z" />
                <path d="M5.25 5.25a3 3 0 0 0-3 3v10.5a3 3 0 0 0 3 3h10.5a3 3 0 0 0 3-3V13.5a.75.75 0 0 0-1.5 0v5.25a1.5 1.5 0 0 1-1.5 1.5H5.25a1.5 1.5 0 0 1-1.5-1.5V8.25a1.5 1.5 0 0 1 1.5-1.5h5.25a.75.75 0 0 0 0-1.5H5.25Z" />
              </svg>
            </button>
          )}
          {hasSavedDraft && (
            <button 
              type="button"
              onClick={loadDraft}
              className="chat-load-draft-button"
              title="Load saved draft"
            >
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" width="16" height="16">
                <path fillRule="evenodd" d="M5.625 1.5H9a3.75 3.75 0 0 1 3.75 3.75v1.875c0 1.036.84 1.875 1.875 1.875H16.5a3.75 3.75 0 0 1 3.75 3.75v7.875c0 1.035-.84 1.875-1.875 1.875H5.625a1.875 1.875 0 0 1-1.875-1.875V3.375c0-1.036.84-1.875 1.875-1.875Zm6.905 9.97a.75.75 0 0 0-1.06 0l-3 3a.75.75 0 1 0 1.06 1.06l1.72-1.72V18a.75.75 0 0 0 1.5 0v-4.19l1.72 1.72a.75.75 0 1 0 1.06-1.06l-3-3Z" clipRule="evenodd" />
                <path d="M14.25 5.25a5.23 5.23 0 0 0-1.279-3.434 9.768 9.768 0 0 1 6.963 6.963A5.23 5.23 0 0 0 16.5 7.5h-1.875a.375.375 0 0 1-.375-.375V5.25Z" />
              </svg>
            </button>
          )}
        </div>
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