import React, { useState, useEffect, useCallback } from "react";
import { createContext, useContext } from "react";
import type { Post } from "../types/Post";
import { useAuth } from "./AuthContext";

interface ISSEContext {
  posts: Post[];
}

const baseUrl = import.meta.env.VITE_BASE_URL;

export const SSEContext = createContext<ISSEContext>(null!);

export function SSEProvider({ children }: { children: React.ReactNode }) {
  const { isAuthenticated } = useAuth();

  const [posts, setPosts] = useState<Post[]>([]);

  const fetchAllPosts = useCallback(async () => {
    try {
      const response = await fetch("/posts");
      if (!response.ok)
        throw new Error(`HTTP error! status: ${response.status}`);
      const data: Post[] = await response.json();
      setPosts(data);
    } catch (e) {
      setPosts([]);
    }
  }, []);

  useEffect(() => {
    if (!isAuthenticated) {
      setPosts([]);
      return;
    }

    fetchAllPosts();

    // Open SSE connection
    const eventSource = new EventSource(`${baseUrl}/events`, {
      withCredentials: true,
    });

    const handleNewPost = (event: MessageEvent) => {
      const newPost = JSON.parse(event.data) as Post;
      setPosts((prevPosts) =>
        prevPosts.some((p) => p.id === newPost.id)
          ? prevPosts
          : [newPost, ...prevPosts]
      );
    };

    const handleUpdatePost = (event: MessageEvent) => {
      const updatedPost = JSON.parse(event.data) as Post;
      setPosts((prevPosts) =>
        prevPosts.map((post) =>
          post.id === updatedPost.id ? updatedPost : post
        )
      );
    };

    const handleDeletePost = (event: MessageEvent) => {
      const { id: deletedPostId } = JSON.parse(event.data) as { id: string };
      setPosts((prevPosts) =>
        prevPosts.filter((post) => post.id !== deletedPostId)
      );
    };

    eventSource.addEventListener("new_post", handleNewPost);
    eventSource.addEventListener("update_post", handleUpdatePost);
    eventSource.addEventListener("delete_post", handleDeletePost);

    eventSource.onerror = (err) => {
      console.error("EventSource failed:", err);
      eventSource.close();
    };

    return () => {
      // Remove listeners and close SSE connection and when user logs out
      eventSource.removeEventListener("new_post", handleNewPost);
      eventSource.removeEventListener("update_post", handleUpdatePost);
      eventSource.removeEventListener("delete_post", handleDeletePost);
      eventSource.close();
    };
  }, [isAuthenticated]);

  const value = { posts };

  return <SSEContext.Provider value={value}>{children}</SSEContext.Provider>;
}

export function useSSE() {
  return useContext(SSEContext);
}
