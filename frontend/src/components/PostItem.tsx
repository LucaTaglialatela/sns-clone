import React, { useState, useRef, useEffect, useCallback } from "react";
import type { Post } from "../types/Post";
import Comments from "../assets/comment.svg";
import Edit from "../assets/edit.svg";
import Delete from "../assets/delete.svg";
import EditPostForm from "./EditPostForm";
import type { Comment } from "../types/Comment";
import { useAuth } from "../context/AuthContext";
import CommentForm from "./CommentForm";
import CommentItem from "./CommentItem";

const COLLAPSED_HEIGHT = 500;

interface PostItemProps {
  post: Post;
  editingPostId: string | null;
  onStartEdit: (post: Post) => void;
  onCancelEdit: () => void;
  onPostUpdated: () => void;
  onDelete: (userId: string, postId: string) => void;
  onViewProfile: (userId: string) => Promise<void>;
}

export const PostItem: React.FC<PostItemProps> = ({
  post,
  editingPostId,
  onStartEdit,
  onCancelEdit,
  onPostUpdated,
  onDelete,
  onViewProfile,
}) => {
  const { userId } = useAuth();
  const [isViewComments, setIsViewComments] = useState<boolean>(false);
  const [comments, setComments] = useState<Comment[]>([]);
  const [isExpanded, setIsExpanded] = useState<boolean>(false);
  const [isOverflowing, setIsOverflowing] = useState<boolean>(false);
  const [text, setText] = useState<string>("");
  const [isSubmitting, setIsSubmitting] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const contentRef = useRef<HTMLDivElement | null>(null);

  const checkOverflow = useCallback(() => {
    const element = contentRef.current;
    if (element) {
      const hasOverflow = element.scrollHeight > COLLAPSED_HEIGHT;
      setIsOverflowing(hasOverflow);
    }
  }, []);

  useEffect(() => {
    fetchPostComments();
  }, []);

  useEffect(() => {
    // When the post content changes, reset to a collapsed state
    // and re-evaluate the overflow
    setIsExpanded(false);

    // Add a small delay to allow for browser layout calculations
    const timeoutId = setTimeout(checkOverflow, 1);
    return () => clearTimeout(timeoutId);
  }, [checkOverflow, post.text, post.image_url]);

  const fetchPostComments = useCallback(async () => {
    try {
      const response = await fetch(`/posts/${post.id}/comments`);
      if (!response.ok)
        throw new Error(`HTTP error! status: ${response.status}`);
      const data: Comment[] = await response.json();
      setComments(data);
    } catch (e) {
      setComments([]);
    }
  }, []);

  const handleCreateComment = async (
    event: React.FormEvent<HTMLFormElement>
  ) => {
    event.preventDefault();

    // Remove invisible characters from the input string
    const invisibleCharRegex =
      /[\u0000-\u0009\u000B-\u000C\u000E-\u001F\u00A0\u115F\u1160\u2000-\u200D\u202A-\u202F\u205F\u2060\u3000\u3164\uFEFF]/g;
    const filteredText = text.replace(invisibleCharRegex, "");

    if (!filteredText.trim()) {
      return;
    }

    setIsSubmitting(true);
    setError(null);

    try {
      const createPostResponse = await fetch(`/posts/${post.id}/comments`, {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ text }),
      });

      if (!createPostResponse.ok) {
        throw new Error(
          `Error: ${createPostResponse.status} ${createPostResponse.statusText}`
        );
      }

      // After creating a post, we empty the text field and file input
      // and signal to the parent component that a post was created
      setText("");
      fetchPostComments();
    } catch (err) {
      const errorMessage =
        err instanceof Error ? err.message : "An unknown error occurred.";
      setError(errorMessage);
      console.error("Failed to create post:", errorMessage);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleDeleteComment = useCallback(
    async (userId: string, postId: string, commentId: string) => {
      try {
        const response = await fetch(
          `/users/${userId}/posts/${postId}/comments/${commentId}`,
          {
            method: "DELETE",
          }
        );
        if (!response.ok) {
          throw new Error("Failed to delete comment.");
        }
        fetchPostComments();
      } catch (err) {
        console.error(err);
        alert(
          err instanceof Error ? err.message : "An unknown error occurred."
        );
      }
    },
    [fetchPostComments]
  );

  return (
    <li className="p-4 rounded-lg shadow-lg/10 bg-white">
      {editingPostId === post.id ? (
        <EditPostForm
          post={post}
          onPostUpdated={onPostUpdated}
          onCancel={onCancelEdit}
        />
      ) : (
        <>
          <div
            ref={contentRef}
            className="relative overflow-hidden transition-all duration-300"
            style={{ maxHeight: isExpanded ? "none" : `${COLLAPSED_HEIGHT}px` }}
          >
            <p className="text-gray-800 whitespace-pre-wrap wrap-break-word">
              {post.text}
            </p>
            {post.image_url && post.image_url !== "" && (
              <div className="mt-3">
                <img
                  src={post.image_url}
                  alt="image"
                  className="max-w-full h-auto max-h-[300px] rounded-lg inset-shadow-sm"
                  // Check again for overflow once the image is loaded
                  onLoad={checkOverflow}
                />
              </div>
            )}
            {!isExpanded && isOverflowing && (
              <div className="absolute bottom-0 left-0 w-full h-12 bg-gradient-to-t from-white to-transparent"></div>
            )}
          </div>

          {isOverflowing && (
            <button
              onClick={() => setIsExpanded(!isExpanded)}
              className="mt-2 cursor-pointer text-sm font-medium hover:text-gray-800 hover:underline"
            >
              {isExpanded ? "See Less" : "See More"}
            </button>
          )}

          <div className="flex justify-between items-center mt-2">
            <div className="text-sm text-gray-500">
              <span>
                Posted by{" "}
                <span
                  onClick={() => onViewProfile(post.user_id)}
                  className="cursor-pointer hover:underline"
                >
                  {post.user_name}
                </span>
              </span>
              <div className="relative group inline-block">
                <span className="ml-4">
                  {new Date(post.timestamp).toLocaleString()}{" "}
                  {!!post.edited ? "(edited)" : ""}
                </span>
                {!!post.edited && (
                  <div className="absolute text-center bottom-full left-1/2 -translate-x-1/2 mb-2 p-2 bg-black text-white text-xs rounded-full whitespace-nowrap opacity-0 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none">
                    Edited on {new Date(post.edited).toLocaleString()}
                  </div>
                )}
              </div>
            </div>
            <div className="flex items-center space-x-2">
              <p className="text-sm font-medium">{comments.length}</p>
              <img
                src={Comments}
                alt="comment"
                onClick={() => setIsViewComments(!isViewComments)}
                className="h-5 w-5 cursor-pointer hover:opacity-75 transition-opacity"
              />
              {post.user_id === userId && (
                <>
                  <img
                    src={Edit}
                    alt="edit"
                    onClick={() => onStartEdit(post)}
                    className="h-5 w-5 cursor-pointer hover:opacity-75 transition-opacity"
                  />
                  <img
                    src={Delete}
                    alt="delete"
                    onClick={() => onDelete(post.user_id, post.id)}
                    className="h-5 w-5 cursor-pointer hover:opacity-75 transition-opacity"
                  />
                </>
              )}
            </div>
          </div>
          {isViewComments ? (
            <div className="mt-4">
              <CommentForm
                text={text}
                setText={setText}
                onSubmit={handleCreateComment}
                error={error}
                isSubmitting={isSubmitting}
                submitButtonText={"Comment"}
              />
              {comments.map((comment, idx) => {
                return (
                  <>
                    {idx > 0 && (
                      <div className="mt-2 h-px w-full bg-gradient-to-r from-transparent via-gray-400 to-transparent"></div>
                    )}
                    <CommentItem
                      post={post}
                      comment={comment}
                      onViewProfile={onViewProfile}
                      handleDeleteComment={handleDeleteComment}
                      handleUpdateComment={fetchPostComments}
                    />
                  </>
                );
              })}
            </div>
          ) : (
            <div className="mt-4">
              {comments.slice(0, 3).map((comment, idx) => {
                return (
                  <>
                    {idx > 0 && (
                      <div className="mt-2 h-px w-full bg-gradient-to-r from-transparent via-gray-400 to-transparent"></div>
                    )}
                    <CommentItem
                      post={post}
                      comment={comment}
                      onViewProfile={onViewProfile}
                      handleDeleteComment={handleDeleteComment}
                      handleUpdateComment={fetchPostComments}
                    />
                  </>
                );
              })}
            </div>
          )}
        </>
      )}
    </li>
  );
};

export default PostItem;
