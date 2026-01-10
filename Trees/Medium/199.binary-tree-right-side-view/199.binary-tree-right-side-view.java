/**
 * Definition for a binary tree node.
 * public class TreeNode {
 *     int val;
 *     TreeNode left;
 *     TreeNode right;
 *     TreeNode() {}
 *     TreeNode(int val) { this.val = val; }
 *     TreeNode(int val, TreeNode left, TreeNode right) {
 *         this.val = val;
 *         this.left = left;
 *         this.right = right;
 *     }
 * }
 */
class Solution {
    public List<Integer> rightSideView(TreeNode root) {
        Deque<TreeNode>q=new ArrayDeque<>();
        List<Integer>list=new ArrayList<>();
        if(root==null) return list;
        q.offer(root);
        while(!q.isEmpty()){
            int size=q.size();
            list.add(q.peek().val);
            for(int i=0;i<size;i++){
                TreeNode temp=q.poll();
                if(temp.right!=null) q.offer(temp.right);
                if(temp.left!=null) q.offer(temp.left);
            }
        }
        return list;
    }
}