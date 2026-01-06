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
    public List<List<Integer>> levelOrder(TreeNode root) {
        Deque<TreeNode>q=new ArrayDeque<>();
        List<List<Integer>>res=new ArrayList<>();
        if(root==null) return res;
        q.offer(root);
        while(!q.isEmpty()){
            int size=q.size();
            List<Integer>list=new ArrayList<>(size);
            for(int i=0;i<size;i++){
                TreeNode temp=q.poll();
                list.add(temp.val);
                if(temp.left!=null)
                    q.offer(temp.left);
                if(temp.right!=null)
                    q.offer(temp.right);
            }
            res.add(list);
        }
        return res;
    }
}