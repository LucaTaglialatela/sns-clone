import type { Comment } from "../types/Comment";
import type { Post } from "../types/Post";
import EditCommentForm from "./EditCommentForm";
import Edit from "../assets/edit.svg";
import Delete from "../assets/delete.svg";
import { useAuth } from "../context/AuthContext";
import { useState } from "react";

interface CommentItemProps {
  post: Post;
  comment: Comment;
  onViewProfile: (userId: string) => void;
  handleDeleteComment: (
    userId: string,
    postId: string,
    commentId: string
  ) => void;
  handleUpdateComment: () => void;
}

export const CommentItem: React.FC<CommentItemProps> = ({
  post,
  comment,
  onViewProfile,
  handleDeleteComment,
  handleUpdateComment,
}) => {
  const { userId } = useAuth();
  const [editingComment, setEditingComment] = useState<Comment | null>(null);

  return (
    <>
      {editingComment && editingComment.id === comment.id ? (
        <div className="p-4 mt-2">
          <EditCommentForm
            postId={post.id}
            comment={comment}
            onCommentUpdated={() => {
              setEditingComment(null);
              handleUpdateComment();
            }}
            onCancel={() => setEditingComment(null)}
          />
        </div>
      ) : (
        <div className="p-4 mt-2">
          <p className="text-sm text-gray-800 whitespace-pre-wrap wrap-break-word">
            {comment.text}
          </p>
          <div className="flex justify-between items-center mt-2">
            <div className="text-xs text-gray-500">
              <span>
                Commented by{" "}
                <span
                  onClick={() => onViewProfile(comment.user_id)}
                  className="cursor-pointer hover:underline"
                >
                  {comment.user_name}
                </span>
              </span>
              <div className="relative group inline-block">
                <span className="ml-4">
                  {new Date(comment.timestamp).toLocaleString()}{" "}
                  {!!comment.edited ? "(edited)" : ""}
                </span>
                {!!comment.edited && (
                  <div className="absolute text-center bottom-full left-1/2 -translate-x-1/2 mb-2 p-2 bg-black text-white text-xs rounded-full whitespace-nowrap opacity-0 group-hover:opacity-100 transition-opacity duration-300 pointer-events-none">
                    Edited on {new Date(comment.edited).toLocaleString()}
                  </div>
                )}
              </div>
            </div>
            {comment.user_id === userId && (
              <div className="flex items-center space-x-2">
                <img
                  src={Edit}
                  alt="edit"
                  onClick={() => setEditingComment(comment)}
                  className="h-5 w-5 cursor-pointer hover:opacity-75 transition-opacity"
                />
                <img
                  src={Delete}
                  alt="delete"
                  onClick={() =>
                    handleDeleteComment(comment.user_id, post.id, comment.id)
                  }
                  className="h-5 w-5 cursor-pointer hover:opacity-75 transition-opacity"
                />
              </div>
            )}
          </div>
        </div>
      )}
    </>
  );
};

export default CommentItem;
