interface PostFormProps {
  text: string;
  setText: (text: string) => void;
  onSubmit: (event: React.FormEvent<HTMLFormElement>) => void;
  error: string | null;
  isSubmitting: boolean;
  submitButtonText: string;
  placeholder?: string;
  onCancel?: () => void;
}

const CommentForm: React.FC<PostFormProps> = ({
  text,
  setText,
  onSubmit,
  error,
  isSubmitting,
  submitButtonText,
  placeholder = "Add a comment!",
  onCancel,
}) => {
  const maxLength = 140;

  return (
    <form onSubmit={onSubmit} className="flex flex-col gap-4">
      {error && (
        <div className="text-red-600 p-3 border border-red-300 bg-red-50 rounded-lg">
          {error}
        </div>
      )}
      <div className="flex items-center space-x-2">
        <textarea
          value={text}
          onChange={(e) => {
            const invisibleCharRegex =
              /[\u0000-\u001F\u00A0\u115F\u1160\u2000-\u200D\u2028-\u202F\u205F\u2060\u3000\u3164\uFEFF\n\r]/g;
            const filteredText = e.target.value.replace(invisibleCharRegex, "");
            setText(filteredText);
          }}
          placeholder={placeholder}
          className="w-full p-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-0 focus:border-black focus:border-2 disabled:opacity-70 resize-none whitespace-pre"
          rows={1}
          maxLength={maxLength}
        />
        <div className="text-sm hover:text-gray-800 font-medium whitespace-nowrap">
          {text.length} / {maxLength} characters
        </div>
        {onCancel && (
          <button
            type="button"
            onClick={onCancel}
            disabled={isSubmitting}
            className="px-3 py-1.5 text-sm font-semibold cursor-pointer text-gray-700 bg-gray-200 rounded-full hover:bg-gray-300 disabled:opacity-70"
          >
            Cancel
          </button>
        )}
        <button
          type="submit"
          disabled={!text.trim() || isSubmitting}
          className="cursor-pointer bg-gray-900 hover:bg-gray-700 text-white font-bold text-sm py-1.5 px-3 rounded-full focus:outline-none focus:shadow-outline disabled:bg-gray-400 disabled:cursor-not-allowed"
        >
          {isSubmitting ? "Submitting..." : submitButtonText}
        </button>
      </div>
    </form>
  );
};

export default CommentForm;
