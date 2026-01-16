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
    int count;
    public int countPairs(TreeNode root, int k) {
        count=0;
        dfs(root, k);
        return count;
    }

    List<Integer> dfs(TreeNode node, int k){
        List<Integer> result = new ArrayList<>();

        if(node==null) return result;
        if(node.left==null && node.right==null) return new ArrayList<>(Arrays.asList(1));

        List<Integer> left=dfs(node.left, k);
        List<Integer> right=dfs(node.right, k);
        for(int i:left){
            for(int j:right){
                if(i+j<=k) count++;
            }
        }
        
        for (int l : left) {
            if (l + 1 <= k) result.add(l + 1);
        }
        for (int r : right) {
            if (r + 1 <= k) result.add(r + 1);
        }

        return result;
    }
}