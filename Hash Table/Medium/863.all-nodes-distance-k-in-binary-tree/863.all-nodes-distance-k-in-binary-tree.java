/**
 * Definition for a binary tree node.
 * public class TreeNode {
 *     int val;
 *     TreeNode left;
 *     TreeNode right;
 *     TreeNode(int x) { val = x; }
 * }
 */
class Solution {
    Map<TreeNode, TreeNode>parentMap;
    public List<Integer> distanceK(TreeNode root, TreeNode target, int k) {
        int dist=0;
        parentMap=new HashMap<>();
        Deque<TreeNode>q=new ArrayDeque<>();
        Set<TreeNode>visited=new HashSet<>();
        q.add(target);
        visited.add(target);

        buildParentMap(root, null);
        while(!q.isEmpty()){
            if(dist==k){
                List<Integer>res=new ArrayList<>();
                while(!q.isEmpty()) res.add(q.poll().val);
                return res;
            }

            int size=q.size();
            for(int i=0;i<size;i++){
                TreeNode node=q.poll();
                if(node.left!=null && visited.add(node.left)) q.add(node.left);
                if(node.right!=null && visited.add(node.right)) q.add(node.right);
                TreeNode par=parentMap.get(node);
                if(par!=null && visited.add(par)) q.add(par);
            }
            dist++;
        }

        return new ArrayList<>();
    }

    void buildParentMap(TreeNode node, TreeNode parent){
        if(node==null) return;
        parentMap.put(node, parent);
        buildParentMap(node.left, node);
        buildParentMap(node.right, node);
    }
}